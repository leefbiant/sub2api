package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func setupAPIKeyHandler(adminSvc service.AdminService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := NewAdminAPIKeyHandler(adminSvc)
	router.PUT("/api/v1/admin/api-keys/:id", h.UpdateGroup)
	return router
}

func TestAdminAPIKeyHandler_UpdateGroup_InvalidID(t *testing.T) {
	router := setupAPIKeyHandler(newStubAdminService())
	body := `{"group_id": 2}`

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/api-keys/abc", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "Invalid API key ID")
}

func TestAdminAPIKeyHandler_UpdateGroup_InvalidJSON(t *testing.T) {
	router := setupAPIKeyHandler(newStubAdminService())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/api-keys/10", bytes.NewBufferString(`{bad json`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "Invalid request")
}

func TestAdminAPIKeyHandler_UpdateGroup_KeyNotFound(t *testing.T) {
	router := setupAPIKeyHandler(newStubAdminService())
	body := `{"group_id": 2}`

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/api-keys/999", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	// ErrAPIKeyNotFound maps to 404
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAdminAPIKeyHandler_UpdateGroup_BindGroup(t *testing.T) {
	router := setupAPIKeyHandler(newStubAdminService())
	body := `{"group_id": 2}`

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/api-keys/10", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)

	var data struct {
		APIKey struct {
			ID      int64  `json:"id"`
			GroupID *int64 `json:"group_id"`
		} `json:"api_key"`
		AutoGrantedGroupAccess bool `json:"auto_granted_group_access"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	require.Equal(t, int64(10), data.APIKey.ID)
	require.NotNil(t, data.APIKey.GroupID)
	require.Equal(t, int64(2), *data.APIKey.GroupID)
}

func TestAdminAPIKeyHandler_UpdateGroup_ServiceError(t *testing.T) {
	svc := &failingUpdateGroupService{
		stubAdminService: newStubAdminService(),
		err:              errors.New("internal failure"),
	}
	router := setupAPIKeyHandler(svc)
	body := `{"group_id": 2}`

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/api-keys/10", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

// H2: empty body → group_id is nil → no-op, returns original key
func TestAdminAPIKeyHandler_UpdateGroup_EmptyBody_NoChange(t *testing.T) {
	router := setupAPIKeyHandler(newStubAdminService())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/api-keys/10", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			APIKey struct {
				ID int64 `json:"id"`
			} `json:"api_key"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Equal(t, int64(10), resp.Data.APIKey.ID)
}

// M2: service returns GROUP_NOT_ACTIVE → handler maps to 400
func TestAdminAPIKeyHandler_UpdateGroup_GroupNotActive(t *testing.T) {
	svc := &failingUpdateGroupService{
		stubAdminService: newStubAdminService(),
		err:              infraerrors.BadRequest("GROUP_NOT_ACTIVE", "target group is not active"),
	}
	router := setupAPIKeyHandler(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/api-keys/10", bytes.NewBufferString(`{"group_id": 5}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "GROUP_NOT_ACTIVE")
}

// M2: service returns INVALID_GROUP_ID → handler maps to 400
func TestAdminAPIKeyHandler_UpdateGroup_NegativeGroupID(t *testing.T) {
	svc := &failingUpdateGroupService{
		stubAdminService: newStubAdminService(),
		err:              infraerrors.BadRequest("INVALID_GROUP_ID", "group_id must be non-negative"),
	}
	router := setupAPIKeyHandler(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/api-keys/10", bytes.NewBufferString(`{"group_id": -5}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "INVALID_GROUP_ID")
}

// stubAdminService implements service.AdminService with zero-value responses.
type stubAdminService struct{}

func (s *stubAdminService) ListUsers(ctx context.Context, page, pageSize int, filters service.UserListFilters, sortBy, sortOrder string) ([]service.User, int64, error) {
	return nil, 0, nil
}
func (s *stubAdminService) GetUser(ctx context.Context, id int64) (*service.User, error) { return nil, nil }
func (s *stubAdminService) CreateUser(ctx context.Context, input *service.CreateUserInput) (*service.User, error) { return nil, nil }
func (s *stubAdminService) UpdateUser(ctx context.Context, id int64, input *service.UpdateUserInput) (*service.User, error) { return nil, nil }
func (s *stubAdminService) DeleteUser(ctx context.Context, id int64) error { return nil }
func (s *stubAdminService) UpdateUserBalance(ctx context.Context, userID int64, balance float64, operation, notes string) (*service.User, error) { return nil, nil }
func (s *stubAdminService) GetUserAPIKeys(ctx context.Context, userID int64, page, pageSize int, sortBy, sortOrder string) ([]service.APIKey, int64, error) { return nil, 0, nil }
func (s *stubAdminService) GetUserUsageStats(ctx context.Context, userID int64, period string) (any, error) { return nil, nil }
func (s *stubAdminService) GetUserRPMStatus(ctx context.Context, userID int64) (*service.UserRPMStatus, error) { return nil, nil }
func (s *stubAdminService) GetUserBalanceHistory(ctx context.Context, userID int64, page, pageSize int, codeType string) ([]service.RedeemCode, int64, float64, error) { return nil, 0, 0, nil }
func (s *stubAdminService) BindUserAuthIdentity(ctx context.Context, userID int64, input service.AdminBindAuthIdentityInput) (*service.AdminBoundAuthIdentity, error) { return nil, nil }
func (s *stubAdminService) ListGroups(ctx context.Context, page, pageSize int, platform, status, search string, isExclusive *bool, sortBy, sortOrder string) ([]service.Group, int64, error) { return nil, 0, nil }
func (s *stubAdminService) GetAllGroups(ctx context.Context) ([]service.Group, error) { return nil, nil }
func (s *stubAdminService) GetAllGroupsByPlatform(ctx context.Context, platform string) ([]service.Group, error) { return nil, nil }
func (s *stubAdminService) GetGroup(ctx context.Context, id int64) (*service.Group, error) { return nil, nil }
func (s *stubAdminService) CreateGroup(ctx context.Context, input *service.CreateGroupInput) (*service.Group, error) { return nil, nil }
func (s *stubAdminService) UpdateGroup(ctx context.Context, id int64, input *service.UpdateGroupInput) (*service.Group, error) { return nil, nil }
func (s *stubAdminService) DeleteGroup(ctx context.Context, id int64) error { return nil }
func (s *stubAdminService) GetGroupAPIKeys(ctx context.Context, groupID int64, page, pageSize int) ([]service.APIKey, int64, error) { return nil, 0, nil }
func (s *stubAdminService) GetGroupRateMultipliers(ctx context.Context, groupID int64) ([]service.UserGroupRateEntry, error) { return nil, nil }
func (s *stubAdminService) ClearGroupRateMultipliers(ctx context.Context, groupID int64) error { return nil }
func (s *stubAdminService) BatchSetGroupRateMultipliers(ctx context.Context, groupID int64, entries []service.GroupRateMultiplierInput) error { return nil }
func (s *stubAdminService) ClearGroupRPMOverrides(ctx context.Context, groupID int64) error { return nil }
func (s *stubAdminService) BatchSetGroupRPMOverrides(ctx context.Context, groupID int64, entries []service.GroupRPMOverrideInput) error { return nil }
func (s *stubAdminService) UpdateGroupSortOrders(ctx context.Context, updates []service.GroupSortOrderUpdate) error { return nil }
func (s *stubAdminService) AdminUpdateAPIKeyGroupID(ctx context.Context, keyID int64, groupID *int64) (*service.AdminUpdateAPIKeyGroupIDResult, error) { return nil, nil }
func (s *stubAdminService) ReplaceUserGroup(ctx context.Context, userID, oldGroupID, newGroupID int64) (*service.ReplaceUserGroupResult, error) { return nil, nil }
func (s *stubAdminService) ListAccounts(ctx context.Context, page, pageSize int, platform, accountType, status, search string, groupID int64, privacyMode, sortBy, sortOrder string) ([]service.Account, int64, error) { return nil, 0, nil }
func (s *stubAdminService) GetAccount(ctx context.Context, id int64) (*service.Account, error) { return nil, nil }
func (s *stubAdminService) GetAccountsByIDs(ctx context.Context, ids []int64) ([]*service.Account, error) { return nil, nil }
func (s *stubAdminService) CreateAccount(ctx context.Context, input *service.CreateAccountInput) (*service.Account, error) { return nil, nil }
func (s *stubAdminService) UpdateAccount(ctx context.Context, id int64, input *service.UpdateAccountInput) (*service.Account, error) { return nil, nil }
func (s *stubAdminService) DeleteAccount(ctx context.Context, id int64) error { return nil }
func (s *stubAdminService) RefreshAccountCredentials(ctx context.Context, id int64) (*service.Account, error) { return nil, nil }
func (s *stubAdminService) ClearAccountError(ctx context.Context, id int64) (*service.Account, error) { return nil, nil }
func (s *stubAdminService) SetAccountError(ctx context.Context, id int64, errorMsg string) error { return nil }
func (s *stubAdminService) EnsureOpenAIPrivacy(ctx context.Context, account *service.Account) string { return "" }
func (s *stubAdminService) EnsureAntigravityPrivacy(ctx context.Context, account *service.Account) string { return "" }
func (s *stubAdminService) ForceOpenAIPrivacy(ctx context.Context, account *service.Account) string { return "" }
func (s *stubAdminService) ForceAntigravityPrivacy(ctx context.Context, account *service.Account) string { return "" }
func (s *stubAdminService) SetAccountSchedulable(ctx context.Context, id int64, schedulable bool) (*service.Account, error) { return nil, nil }
func (s *stubAdminService) BulkUpdateAccounts(ctx context.Context, input *service.BulkUpdateAccountsInput) (*service.BulkUpdateAccountsResult, error) { return nil, nil }
func (s *stubAdminService) CheckMixedChannelRisk(ctx context.Context, currentAccountID int64, currentAccountPlatform string, groupIDs []int64) error { return nil }
func (s *stubAdminService) ListProxies(ctx context.Context, page, pageSize int, protocol, status, search, sortBy, sortOrder string) ([]service.Proxy, int64, error) { return nil, 0, nil }
func (s *stubAdminService) ListProxiesWithAccountCount(ctx context.Context, page, pageSize int, protocol, status, search, sortBy, sortOrder string) ([]service.ProxyWithAccountCount, int64, error) { return nil, 0, nil }
func (s *stubAdminService) GetAllProxies(ctx context.Context) ([]service.Proxy, error) { return nil, nil }
func (s *stubAdminService) GetAllProxiesWithAccountCount(ctx context.Context) ([]service.ProxyWithAccountCount, error) { return nil, nil }
func (s *stubAdminService) GetProxy(ctx context.Context, id int64) (*service.Proxy, error) { return nil, nil }
func (s *stubAdminService) GetProxiesByIDs(ctx context.Context, ids []int64) ([]service.Proxy, error) { return nil, nil }
func (s *stubAdminService) CreateProxy(ctx context.Context, input *service.CreateProxyInput) (*service.Proxy, error) { return nil, nil }
func (s *stubAdminService) UpdateProxy(ctx context.Context, id int64, input *service.UpdateProxyInput) (*service.Proxy, error) { return nil, nil }
func (s *stubAdminService) DeleteProxy(ctx context.Context, id int64) error { return nil }
func (s *stubAdminService) BatchDeleteProxies(ctx context.Context, ids []int64) (*service.ProxyBatchDeleteResult, error) { return nil, nil }
func (s *stubAdminService) GetProxyAccounts(ctx context.Context, proxyID int64) ([]service.ProxyAccountSummary, error) { return nil, nil }
func (s *stubAdminService) CheckProxyExists(ctx context.Context, host string, port int, username, password string) (bool, error) { return false, nil }
func (s *stubAdminService) TestProxy(ctx context.Context, id int64) (*service.ProxyTestResult, error) { return nil, nil }
func (s *stubAdminService) CheckProxyQuality(ctx context.Context, id int64) (*service.ProxyQualityCheckResult, error) { return nil, nil }
func (s *stubAdminService) ListRedeemCodes(ctx context.Context, page, pageSize int, codeType, status, search, sortBy, sortOrder string) ([]service.RedeemCode, int64, error) { return nil, 0, nil }
func (s *stubAdminService) GetRedeemCode(ctx context.Context, id int64) (*service.RedeemCode, error) { return nil, nil }
func (s *stubAdminService) GenerateRedeemCodes(ctx context.Context, input *service.GenerateRedeemCodesInput) ([]service.RedeemCode, error) { return nil, nil }
func (s *stubAdminService) DeleteRedeemCode(ctx context.Context, id int64) error { return nil }
func (s *stubAdminService) BatchDeleteRedeemCodes(ctx context.Context, ids []int64) (int64, error) { return 0, nil }
func (s *stubAdminService) ExpireRedeemCode(ctx context.Context, id int64) (*service.RedeemCode, error) { return nil, nil }
func (s *stubAdminService) ResetAccountQuota(ctx context.Context, id int64) error { return nil }

func newStubAdminService() *stubAdminService { return &stubAdminService{} }

// failingUpdateGroupService overrides AdminUpdateAPIKeyGroupID to return an error.
type failingUpdateGroupService struct {
	*stubAdminService
	err error
}

func (f *failingUpdateGroupService) AdminUpdateAPIKeyGroupID(_ context.Context, _ int64, _ *int64) (*service.AdminUpdateAPIKeyGroupIDResult, error) {
	return nil, f.err
}
