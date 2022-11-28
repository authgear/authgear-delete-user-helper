package handler

import (
	"database/sql"
	"errors"
	"net/http"

	relay "github.com/authgear/graphql-go-relay"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var ErrUserNotFound = apierrors.NewNotFound("user not found")

func ConfigureSearchUserRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("POST").
		WithPathPattern("/_api/delete-user-helper/search")
}

var SearchUserInputSchema = validation.NewSimpleSchema(`
{
	"oneOf": [
		{
			"type": "object",
			"additionalProperties": false,
			"properties": {
				"name": { "type": "string" }
			},
			"required": ["name"]
		},
		{
			"type": "object",
			"additionalProperties": false,
			"properties": {
				"provider_user_id": { "type": "string" }
			},
			"required": ["provider_user_id"]
		}
	]
}
`)

type SearchUserInput struct {
	Name           string `json:"name,omitempty"`
	ProviderUserID string `json:"provider_user_id,omitempty"`
}

type SearchUserResult struct {
	UserID string `json:"user_id,omitempty"`
	NodeID string `json:"node_id,omitempty"`
}

type SearchUserHandler struct {
	AppDatabase *appdb.Handle
	JSON        *httputil.JSONResponseWriter
	SQLBuilder  *appdb.SQLBuilderApp
	SQLExecutor *appdb.SQLExecutor
}

func (h *SearchUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var result *SearchUserResult
	err := h.AppDatabase.WithTx(func() (err error) {
		result, err = h.Handle(w, r)
		return err
	})
	if err == nil {
		h.JSON.WriteResponse(w, &api.Response{Result: result})
	} else {
		h.JSON.WriteResponse(w, &api.Response{Error: err})
	}
}

func (h *SearchUserHandler) Handle(w http.ResponseWriter, r *http.Request) (*SearchUserResult, error) {
	var input SearchUserInput
	if err := httputil.BindJSONBody(r, w, SearchUserInputSchema.Validator(), &input); err != nil {
		return nil, err
	}

	return h.search(input)
}

func (h *SearchUserHandler) search(input SearchUserInput) (*SearchUserResult, error) {
	q := h.SQLBuilder.
		Select(
			"b.user_id",
		).
		From(h.SQLBuilder.TableName("_auth_identity_oauth"), "a").
		Join(h.SQLBuilder.TableName("_auth_identity"), "b", "a.id = b.id").
		Where("a.provider_type = ?", string(config.OAuthSSOProviderTypeAzureADB2C))

	if input.ProviderUserID != "" {
		q = q.Where("a.provider_user_id = ?", input.ProviderUserID)
	}
	if input.Name != "" {
		q = q.Where("a.claims->>'name' = ?", input.Name)
	}

	row, err := h.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	var userID string

	err = row.Scan(
		&userID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	return &SearchUserResult{
		UserID: userID,
		NodeID: relay.ToGlobalID("User", userID),
	}, nil
}
