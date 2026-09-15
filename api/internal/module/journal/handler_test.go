package journal_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/xuri/excelize/v2"

	"pengbook/api/internal/middleware"
	"pengbook/api/internal/module/account"
	"pengbook/api/internal/module/journal"
)

// apiResponse is the standard JSON response wrapper.
type apiResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Message string          `json:"message,omitempty"`
}

// authMiddleware creates a middleware that injects userID into context.
func authMiddleware(userID int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := middleware.ContextWithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// createMultipartRequest creates a multipart form request with a CSV file.
func createMultipartRequest(t *testing.T, url, csvContent, filename string) *http.Request {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	if _, err := part.Write([]byte(csvContent)); err != nil {
		t.Fatalf("write file content: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestHandler_GetAllScrollView(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	// Create a journal entry first
	entry := &journal.JournalEntry{
		UserID:      userID,
		Date:        time.Now(),
		Description: "Handler test entry",
		Lines: []journal.JournalEntryLine{
			{AccountID: accID1, Debit: 100000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 100000},
		},
	}
	if err := tc.jnlRepo.CreateEntry(ctx, entry); err != nil {
		t.Fatalf("CreateEntry: %v", err)
	}
	defer tc.jnlRepo.DeleteEntry(ctx, entry.ID)

	// Setup router with handler and auth middleware
	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/journals?limit=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success=true, got message: %s", resp.Message)
	}
}

func TestHandler_GetAllScrollView_Unauthorized(t *testing.T) {
	tc := setup(t)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	// No auth middleware
	r.Mount("/api/v1/journals", handler.Routes())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/journals", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_GetTotalSummary(t *testing.T) {
	tc := setup(t)
	userID := createUser(t, tc)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/journals/summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success=true, got message: %s", resp.Message)
	}
}

func TestHandler_Create_Success(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	body := journal.CreateJournalRequest{
		Date:        time.Now().Format(time.RFC3339),
		Description: "Handler create test",
		Lines: []journal.JournalLineDto{
			{AccountID: accID1, Debit: 100000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 100000},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/journals", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success=true, got message: %s", resp.Message)
	}

	// Cleanup: find and delete the created entry
	entries, _ := tc.jnlRepo.FindEntriesByUserID(ctx, userID, journal.EntryFilter{Limit: 10})
	for _, e := range entries {
		if e.Description == "Handler create test" {
			tc.jnlRepo.DeleteEntry(ctx, e.ID)
			break
		}
	}
}

func TestHandler_Create_Unbalanced(t *testing.T) {
	tc := setup(t)
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	body := journal.CreateJournalRequest{
		Date:        time.Now().Format(time.RFC3339),
		Description: "Unbalanced entry",
		Lines: []journal.JournalLineDto{
			{AccountID: accID1, Debit: 200000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 100000},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/journals", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Create_InvalidBody(t *testing.T) {
	tc := setup(t)
	userID := createUser(t, tc)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/journals", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Create_InvalidLines(t *testing.T) {
	tc := setup(t)
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	body := journal.CreateJournalRequest{
		Date:        time.Now().Format(time.RFC3339),
		Description: "Only one line",
		Lines: []journal.JournalLineDto{
			{AccountID: accID1, Debit: 100000, Credit: 0},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/journals", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Update_Success(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	// Create entry first
	entry := &journal.JournalEntry{
		UserID:      userID,
		Date:        time.Now(),
		Description: "Original",
		Lines: []journal.JournalEntryLine{
			{AccountID: accID1, Debit: 100000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 100000},
		},
	}
	if err := tc.jnlRepo.CreateEntry(ctx, entry); err != nil {
		t.Fatalf("CreateEntry: %v", err)
	}
	defer tc.jnlRepo.DeleteEntry(ctx, entry.ID)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	body := journal.UpdateJournalRequest{
		Date:        time.Now().Format(time.RFC3339),
		Description: "Updated",
		Lines: []journal.JournalLineDto{
			{AccountID: accID1, Debit: 200000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 200000},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	url := "/api/v1/journals/" + strconv.FormatInt(entry.ID, 10)
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success=true, got message: %s", resp.Message)
	}
}

func TestHandler_Update_NotFound(t *testing.T) {
	tc := setup(t)
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	body := journal.UpdateJournalRequest{
		Date:        time.Now().Format(time.RFC3339),
		Description: "Not found",
		Lines: []journal.JournalLineDto{
			{AccountID: accID1, Debit: 100000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 100000},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/journals/999999", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Upload_Success(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	// Get account codes
	acc1, _ := tc.accRepo.FindByID(ctx, accID1)
	acc2, _ := tc.accRepo.FindByID(ctx, accID2)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	// Create CSV content
	date := time.Now().Format("2006-01-02")
	csvContent := fmt.Sprintf(`Tanggal,Deskripsi,Kode Akun,Debit,Kredit
%s,Upload test,%s,100000,0
%s,Upload test,%s,0,100000`, date, acc1.Code, date, acc2.Code)

	req := createMultipartRequest(t, "/api/v1/journals/upload", csvContent, "test.csv")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success=true, got message: %s", resp.Message)
	}

	// Cleanup
	entries, _ := tc.jnlRepo.FindEntriesByUserID(ctx, userID, journal.EntryFilter{Limit: 100})
	for _, e := range entries {
		if e.Description == "Upload test" {
			tc.jnlRepo.DeleteEntry(ctx, e.ID)
		}
	}
}

func TestHandler_Upload_MultipleEntries(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	acc1, _ := tc.accRepo.FindByID(ctx, accID1)
	acc2, _ := tc.accRepo.FindByID(ctx, accID2)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	// CSV with 2 different entries (different dates)
	csvContent := fmt.Sprintf(`Tanggal,Deskripsi,Kode Akun,Debit,Kredit
2026-04-21,Entry 1,%s,100000,0
2026-04-21,Entry 1,%s,0,100000
2026-04-22,Entry 2,%s,200000,0
2026-04-22,Entry 2,%s,0,200000`, acc1.Code, acc2.Code, acc1.Code, acc2.Code)

	req := createMultipartRequest(t, "/api/v1/journals/upload", csvContent, "test.csv")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success=true, got message: %s", resp.Message)
	}

	// Cleanup
	entries, _ := tc.jnlRepo.FindEntriesByUserID(ctx, userID, journal.EntryFilter{Limit: 100})
	for _, e := range entries {
		if strings.HasPrefix(e.Description, "Entry ") {
			tc.jnlRepo.DeleteEntry(ctx, e.ID)
		}
	}
}

func TestHandler_Upload_Unbalanced(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	acc1, _ := tc.accRepo.FindByID(ctx, accID1)
	acc2, _ := tc.accRepo.FindByID(ctx, accID2)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	// Unbalanced entry (debit != credit)
	date := time.Now().Format("2006-01-02")
	csvContent := fmt.Sprintf(`Tanggal,Deskripsi,Kode Akun,Debit,Kredit
%s,Unbalanced,%s,200000,0
%s,Unbalanced,%s,0,100000`, date, acc1.Code, date, acc2.Code)

	req := createMultipartRequest(t, "/api/v1/journals/upload", csvContent, "test.csv")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Upload_InvalidCSV_NotPostingAccount(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)

	// Create a header account (level < 3)
	headerAcc, _ := tc.accSvc.Create(ctx, userID, account.CreateAccountRequest{
		Code: "1.01.00.00",
		Name: "Header Account",
	})
	accID2 := createAccount(t, tc, userID, "4.01.01.01")
	acc2, _ := tc.accRepo.FindByID(ctx, accID2)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	// Use non-posting account
	date := time.Now().Format("2006-01-02")
	csvContent := fmt.Sprintf(`Tanggal,Deskripsi,Kode Akun,Debit,Kredit
%s,Test,%s,100000,0
%s,Test,%s,0,100000`, date, headerAcc.Code, date, acc2.Code)

	req := createMultipartRequest(t, "/api/v1/journals/upload", csvContent, "test.csv")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Upload_NoFile(t *testing.T) {
	tc := setup(t)
	userID := createUser(t, tc)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	// Send request without file
	req := httptest.NewRequest(http.MethodPost, "/api/v1/journals/upload", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Upload_WrongExtension(t *testing.T) {
	tc := setup(t)
	userID := createUser(t, tc)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	// Send .txt file instead of .csv
	req := createMultipartRequest(t, "/api/v1/journals/upload", "some content", "test.txt")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Upload_EmptyCSV(t *testing.T) {
	tc := setup(t)
	userID := createUser(t, tc)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	// Send empty CSV
	csvContent := `Tanggal,Deskripsi,Kode Akun,Debit,Kredit`
	req := createMultipartRequest(t, "/api/v1/journals/upload", csvContent, "empty.csv")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Upload_InvalidCSVFormat(t *testing.T) {
	tc := setup(t)
	userID := createUser(t, tc)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	// CSV with missing required column
	csvContent := `Tanggal,Deskripsi,Kode Akun
2026-04-21,Test,1.01.01.01`
	req := createMultipartRequest(t, "/api/v1/journals/upload", csvContent, "invalid.csv")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Upload_Unauthorized(t *testing.T) {
	tc := setup(t)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	// No auth middleware
	r.Mount("/api/v1/journals", handler.Routes())

	csvContent := `Tanggal,Deskripsi,Kode Akun,Debit,Kredit
2026-04-21,Test,1.01.01.01,100000,0`
	req := createMultipartRequest(t, "/api/v1/journals/upload", csvContent, "test.csv")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}

// ── DownloadTemplate Tests ──────────────────────────────────────────────────

func TestHandler_DownloadTemplate_Success(t *testing.T) {
	tc := setup(t)
	userID := createUser(t, tc)
	// Create some posting accounts
	createAccount(t, tc, userID, "1.01.01.01")
	createAccount(t, tc, userID, "4.01.01.01")

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/journals/template", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "spreadsheetml") {
		t.Errorf("expected Excel content type, got %s", contentType)
	}

	// Check content disposition
	disposition := w.Header().Get("Content-Disposition")
	if !strings.Contains(disposition, "template-jurnal.xlsx") {
		t.Errorf("expected attachment filename, got %s", disposition)
	}

	// Check body is not empty
	if w.Body.Len() == 0 {
		t.Fatal("expected non-empty response body")
	}
}

func TestHandler_DownloadTemplate_Unauthorized(t *testing.T) {
	tc := setup(t)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	// No auth middleware
	r.Mount("/api/v1/journals", handler.Routes())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/journals/template", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_DownloadTemplate_NoAccounts(t *testing.T) {
	tc := setup(t)
	userID := createUser(t, tc)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/journals/template", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should still return a template (with example accounts)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ── Upload Excel Tests ─────────────────────────────────────────────────────

func TestHandler_Upload_Excel_Success(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	// Get account codes
	acc1, _ := tc.accRepo.FindByID(ctx, accID1)
	acc2, _ := tc.accRepo.FindByID(ctx, accID2)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	// Create Excel file
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"
	f.SetSheetName("Sheet1", sheet)

	// Header
	f.SetCellValue(sheet, "A1", "Tanggal")
	f.SetCellValue(sheet, "B1", "Deskripsi")
	f.SetCellValue(sheet, "C1", "Kode Akun")
	f.SetCellValue(sheet, "D1", "Debit")
	f.SetCellValue(sheet, "E1", "Kredit")

	// Data rows
	date := time.Now().Format("2006-01-02")
	f.SetCellValue(sheet, "A2", date)
	f.SetCellValue(sheet, "B2", "Upload Excel test")
	f.SetCellValue(sheet, "C2", acc1.Code)
	f.SetCellValue(sheet, "D2", 100000)
	f.SetCellValue(sheet, "E2", 0)

	f.SetCellValue(sheet, "A3", date)
	f.SetCellValue(sheet, "B3", "Upload Excel test")
	f.SetCellValue(sheet, "C3", acc2.Code)
	f.SetCellValue(sheet, "D3", 0)
	f.SetCellValue(sheet, "E3", 100000)

	buffer, err := f.WriteToBuffer()
	if err != nil {
		t.Fatalf("failed to create Excel: %v", err)
	}

	// Create multipart request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.xlsx")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	part.Write(buffer.Bytes())
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/journals/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success=true, got message: %s", resp.Message)
	}

	// Cleanup
	entries, _ := tc.jnlRepo.FindEntriesByUserID(ctx, userID, journal.EntryFilter{Limit: 100})
	for _, e := range entries {
		if e.Description == "Upload Excel test" {
			tc.jnlRepo.DeleteEntry(ctx, e.ID)
		}
	}
}

func TestHandler_Upload_WrongExtension_Excel(t *testing.T) {
	tc := setup(t)
	userID := createUser(t, tc)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	// Send .txt file
	req := createMultipartRequest(t, "/api/v1/journals/upload", "some content", "test.txt")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Delete_Success(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	// Create entry via service so balance deltas are applied
	created, err := tc.jnlSvc.Create(ctx, userID, journal.CreateJournalRequest{
		Date:        time.Now().Format(time.RFC3339),
		Description: "To be deleted",
		Lines: []journal.JournalLineDto{
			{AccountID: accID1, Debit: 100000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 100000},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	entryID := created.ID

	getBalance := func(accountID int64) float64 {
		t.Helper()
		var balance float64
		if err := tc.pool.QueryRow(ctx,
			"SELECT COALESCE(balance, 0) FROM account_balances WHERE account_id = $1",
			accountID).Scan(&balance); err != nil {
			t.Fatalf("query balance: %v", err)
		}
		return balance
	}

	// Balances must reflect the created entry
	if got := getBalance(accID1); got != 100000 {
		t.Fatalf("expected balance 100000 for acc %d after create, got %v", accID1, got)
	}
	if got := getBalance(accID2); got != 100000 {
		t.Fatalf("expected balance 100000 for acc %d after create, got %v", accID2, got)
	}

	type auditRow struct {
		action  string
		entryID *int64
	}
	queryAudits := func() []auditRow {
		t.Helper()
		rows, err := tc.pool.Query(ctx,
			"SELECT action, journal_entry_id FROM journal_audit_logs WHERE user_id = $1 ORDER BY id",
			userID)
		if err != nil {
			t.Fatalf("query audit logs: %v", err)
		}
		defer rows.Close()

		var audits []auditRow
		for rows.Next() {
			var a auditRow
			if err := rows.Scan(&a.action, &a.entryID); err != nil {
				t.Fatalf("scan audit log: %v", err)
			}
			audits = append(audits, a)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("audit rows: %v", err)
		}
		return audits
	}

	// Created audit must link to the entry while it still exists.
	if audits := queryAudits(); len(audits) != 1 ||
		audits[0].action != "journal.created" ||
		audits[0].entryID == nil || *audits[0].entryID != entryID {
		t.Fatalf("expected 1 journal.created audit for entry %d, got %+v", entryID, audits)
	}

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	url := "/api/v1/journals/" + strconv.FormatInt(entryID, 10)
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success=true, got message: %s", resp.Message)
	}

	// Entry must be gone
	gone, err := tc.jnlRepo.FindEntryByID(ctx, entryID)
	if err != nil {
		t.Fatalf("FindEntryByID: %v", err)
	}
	if gone != nil {
		t.Fatalf("expected entry to be deleted, still exists: %d", entryID)
	}

	// Balances must be reversed to zero
	if got := getBalance(accID1); got != 0 {
		t.Fatalf("expected balance 0 for acc %d after delete, got %v", accID1, got)
	}
	if got := getBalance(accID2); got != 0 {
		t.Fatalf("expected balance 0 for acc %d after delete, got %v", accID2, got)
	}

	// Audit trail must contain created + deleted for this user, in order.
	// NOTE: deleting the entry fires ON DELETE SET NULL, so BOTH rows end up
	// with NULL entry ref (same convention as accounts_audit_logs) — the
	// old_values JSONB snapshots keep the actual data.
	audits := queryAudits()
	if len(audits) != 2 {
		t.Fatalf("expected 2 audit rows, got %d: %+v", len(audits), audits)
	}
	if audits[0].action != "journal.created" {
		t.Fatalf("expected first audit journal.created, got %s", audits[0].action)
	}
	if audits[1].action != "journal.deleted" {
		t.Fatalf("expected second audit journal.deleted, got %s", audits[1].action)
	}
}

func TestHandler_Delete_NotFound(t *testing.T) {
	tc := setup(t)
	userID := createUser(t, tc)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/journals/999999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Delete_Unauthorized(t *testing.T) {
	tc := setup(t)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	// No auth middleware
	r.Mount("/api/v1/journals", handler.Routes())

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/journals/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	tc := setup(t)
	userID := createUser(t, tc)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/journals/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Export_Success(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")
	acc1, _ := tc.accRepo.FindByID(ctx, accID1)

	created, err := tc.jnlSvc.Create(ctx, userID, journal.CreateJournalRequest{
		Date:        time.Now().Format(time.RFC3339),
		Description: "Export me",
		Lines: []journal.JournalLineDto{
			{AccountID: accID1, Debit: 100000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 100000},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer tc.jnlRepo.DeleteEntry(ctx, created.ID)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/journals/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "spreadsheetml") {
		t.Errorf("expected Excel content type, got %s", contentType)
	}

	if w.Body.Len() == 0 {
		t.Fatal("expected non-empty response body")
	}

	// Exported file must be re-uploadable and contain our entry.
	parsed, err := journal.ParseExcel(bytes.NewReader(w.Body.Bytes()), time.UTC)
	if err != nil {
		t.Fatalf("ParseExcel(exported): %v", err)
	}
	if len(parsed) != 1 {
		t.Fatalf("expected 1 entry in export, got %d", len(parsed))
	}
	if parsed[0].Description != "Export me" {
		t.Errorf("expected description kept, got %q", parsed[0].Description)
	}
	if parsed[0].Lines[0].AccountCode != acc1.Code {
		t.Errorf("expected account code %s, got %s", acc1.Code, parsed[0].Lines[0].AccountCode)
	}
}

func TestHandler_Summary_WithMonth_Success(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accKas := createAccount(t, tc, userID, "1.01.01.01")
	accPendapatan := createAccount(t, tc, userID, "4.01.01.01")
	accBeban := createAccount(t, tc, userID, "5.01.01.01")

	create := func(date, desc string, lines []journal.JournalLineDto) {
		t.Helper()
		if _, err := tc.jnlSvc.Create(ctx, userID, journal.CreateJournalRequest{
			Date: date, Description: desc, Lines: lines,
		}); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	// March: income 500k, expense 200k, 2 transactions
	create("2026-03-10T12:00:00+07:00", "Penjualan", []journal.JournalLineDto{
		{AccountID: accKas, Debit: 500000, Credit: 0},
		{AccountID: accPendapatan, Debit: 0, Credit: 500000},
	})
	create("2026-03-15T12:00:00+07:00", "Beban operasional", []journal.JournalLineDto{
		{AccountID: accBeban, Debit: 200000, Credit: 0},
		{AccountID: accKas, Debit: 0, Credit: 200000},
	})
	// April: must be excluded from the March summary
	create("2026-04-05T12:00:00+07:00", "Penjualan April", []journal.JournalLineDto{
		{AccountID: accKas, Debit: 999000, Credit: 0},
		{AccountID: accPendapatan, Debit: 0, Credit: 999000},
	})

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	// Full ISO instant with client offset; backend derives March bounds in +07:00.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/journals/summary?month=2026-03-10T12%3A00%3A00%2B07%3A00", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success=true, got message: %s", resp.Message)
	}

	var summary struct {
		Month            string  `json:"month"`
		Income           float64 `json:"income"`
		Expense          float64 `json:"expense"`
		TransactionCount int64   `json:"transactionCount"`
	}
	if err := json.Unmarshal(resp.Data, &summary); err != nil {
		t.Fatalf("unmarshal summary: %v", err)
	}

	// No all-time fields may leak into the response.
	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Data, &raw); err != nil {
		t.Fatalf("unmarshal raw summary: %v", err)
	}
	for _, key := range []string{"totalDebit", "totalCredit"} {
		if _, exists := raw[key]; exists {
			t.Errorf("expected no %q field in summary response", key)
		}
	}

	// Monthly slice covers March only.
	if summary.Month != "2026-03" {
		t.Errorf("expected month 2026-03, got %s", summary.Month)
	}
	if summary.Income != 500000 {
		t.Errorf("expected income 500000, got %v", summary.Income)
	}
	if summary.Expense != 200000 {
		t.Errorf("expected expense 200000, got %v", summary.Expense)
	}
	if summary.TransactionCount != 2 {
		t.Errorf("expected transactionCount 2, got %d", summary.TransactionCount)
	}

	// Cleanup
	entries, _ := tc.jnlRepo.FindEntriesByUserID(ctx, userID, journal.EntryFilter{Limit: 100})
	for _, e := range entries {
		tc.jnlRepo.DeleteEntry(ctx, e.ID)
	}
}

func TestHandler_Summary_ClientTimezone(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accKas := createAccount(t, tc, userID, "1.01.01.01")
	accPendapatan := createAccount(t, tc, userID, "4.01.01.01")

	// Entry on 2026-03-10 noon Jakarta.
	if _, err := tc.jnlSvc.Create(ctx, userID, journal.CreateJournalRequest{
		Date:        "2026-03-10T12:00:00+07:00",
		Description: "Penjualan",
		Lines: []journal.JournalLineDto{
			{AccountID: accKas, Debit: 500000, Credit: 0},
			{AccountID: accPendapatan, Debit: 0, Credit: 500000},
		},
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	get := func(monthParam string) map[string]interface{} {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/journals/summary?month="+monthParam, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("month %s: expected status 200, got %d: %s", monthParam, w.Code, w.Body.String())
		}
		var resp apiResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		var summary map[string]interface{}
		if err := json.Unmarshal(resp.Data, &summary); err != nil {
			t.Fatalf("unmarshal summary: %v", err)
		}
		return summary
	}

	// Same wall-clock intent, different client zones.
	jakarta := get("2026-03-10T12%3A00%3A00%2B07%3A00")
	if jakarta["month"] != "2026-03" || jakarta["income"] != 500000.0 {
		t.Errorf("jakarta offset: expected 2026-03/500000, got %+v", jakarta)
	}

	// 2026-10-01 00:30 +07:00 is still 2026-09-30 in UTC: a UTC-based parser
	// would wrongly return September (with data). Offset-aware parsing must
	// return October (empty).
	october := get("2026-10-01T00%3A30%3A00%2B07%3A00")
	if october["month"] != "2026-10" {
		t.Errorf("expected month 2026-10, got %+v", october)
	}
	if october["income"] != 0.0 || october["transactionCount"] != 0.0 {
		t.Errorf("expected empty October slice, got %+v", october)
	}

	// Cleanup
	entries, _ := tc.jnlRepo.FindEntriesByUserID(ctx, userID, journal.EntryFilter{Limit: 100})
	for _, e := range entries {
		tc.jnlRepo.DeleteEntry(ctx, e.ID)
	}
}

func TestHandler_Summary_InvalidMonth(t *testing.T) {
	tc := setup(t)
	userID := createUser(t, tc)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/journals/summary?month=bukan-bulan", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Summary_Month_Unauthorized(t *testing.T) {
	tc := setup(t)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	// No auth middleware
	r.Mount("/api/v1/journals", handler.Routes())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/journals/summary?month=2026-03", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_GetAllScrollView_InvalidFilter(t *testing.T) {
	tc := setup(t)
	userID := createUser(t, tc)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	for _, query := range []string{
		"limit=10&startDate=bukan-tanggal",
		"limit=10&endDate=bukan-tanggal",
		"limit=10&cursor=bukan-cursor",
		"limit=10&cursor=2026-09-16T05%3A00%3A00.000Z_bukan-id",
	} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/journals?"+query, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("query %q: expected status 400, got %d: %s", query, w.Code, w.Body.String())
		}
	}
}

func TestHandler_Export_Timezone(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	// 2026-03-10 02:00 +07:00 == 2026-03-09 19:00Z == 2026-03-10 08:00 +13:00.
	// UTC and Auckland (+13:00 DST) disagree on the calendar day: the tz
	// param decides which one the export renders.
	created, err := tc.jnlSvc.Create(ctx, userID, journal.CreateJournalRequest{
		Date:        "2026-03-10T02:00:00+07:00",
		Description: "TZ export",
		Lines: []journal.JournalLineDto{
			{AccountID: accID1, Debit: 100000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 100000},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer tc.jnlRepo.DeleteEntry(ctx, created.ID)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Use(middleware.Timezone)
	r.Mount("/api/v1/journals", handler.Routes())

	get := func(tz string) []byte {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/journals/export", nil)
		if tz != "" {
			req.Header.Set("X-Timezone", tz)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("tz %q: expected status 200, got %d: %s", tz, w.Code, w.Body.String())
		}
		return w.Body.Bytes()
	}

	// Auckland (+13:00 DST in March) renders 2026-03-10...
	auckland, err := journal.ParseExcel(bytes.NewReader(get("Pacific/Auckland")), time.UTC)
	if err != nil {
		t.Fatalf("ParseExcel(auckland export): %v", err)
	}
	if len(auckland) != 1 || !strings.HasPrefix(auckland[0].Date, "2026-03-10") {
		t.Fatalf("expected 2026-03-10 entry in Auckland export, got %+v", auckland)
	}

	// ...while UTC renders the previous calendar day for the same instant.
	// This pair proves the header actually drives the rendering.
	utc, err := journal.ParseExcel(bytes.NewReader(get("")), time.UTC)
	if err != nil {
		t.Fatalf("ParseExcel(utc export): %v", err)
	}
	if len(utc) != 1 || !strings.HasPrefix(utc[0].Date, "2026-03-09") {
		t.Fatalf("expected 2026-03-09 entry in UTC export, got %+v", utc)
	}

	// Jakarta keeps the original calendar day.
	jakarta, err := journal.ParseExcel(bytes.NewReader(get("Asia/Jakarta")), time.UTC)
	if err != nil {
		t.Fatalf("ParseExcel(jakarta export): %v", err)
	}
	if len(jakarta) != 1 || !strings.HasPrefix(jakarta[0].Date, "2026-03-10") {
		t.Fatalf("expected 2026-03-10 entry in Jakarta export, got %+v", jakarta)
	}

	// Unknown zone falls back to UTC instead of failing.
	get("Bukan/Zona")

	// Invalid date filter is a 400, not a silent unfiltered export.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/journals/export?startDate=bogus", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Export_Unauthorized(t *testing.T) {
	tc := setup(t)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	// No auth middleware
	r.Mount("/api/v1/journals", handler.Routes())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/journals/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Upload_EmptyExcel(t *testing.T) {
	tc := setup(t)
	userID := createUser(t, tc)

	handler := journal.NewHandler(tc.jnlSvc)
	r := chi.NewRouter()
	r.Use(authMiddleware(userID))
	r.Mount("/api/v1/journals", handler.Routes())

	// Create empty Excel
	f := excelize.NewFile()
	defer f.Close()
	buffer, _ := f.WriteToBuffer()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "empty.xlsx")
	part.Write(buffer.Bytes())
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/journals/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}
