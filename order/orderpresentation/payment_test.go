package orderpresentation

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSavePaymentProofAcceptsPDF(t *testing.T) {
	t.Parallel()

	handler, ctx := newPaymentProofTestContext(t, "receipt.pdf", []byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj\n"))

	proof, err := handler.savePaymentProof(ctx, "payment_proof")
	if err != nil {
		t.Fatalf("savePaymentProof returned error: %v", err)
	}

	if !strings.HasPrefix(proof.PublicPath, paymentProofPathPrefix) {
		t.Fatalf("expected public path to start with %q, got %q", paymentProofPathPrefix, proof.PublicPath)
	}
	if filepath.Dir(proof.SavedPath) != handler.paymentProofDir {
		t.Fatalf("expected saved path inside %q, got %q", handler.paymentProofDir, proof.SavedPath)
	}
	if _, err := os.Stat(proof.SavedPath); err != nil {
		t.Fatalf("expected saved file to exist: %v", err)
	}
}

func TestSavePaymentProofRejectsMismatchedContentType(t *testing.T) {
	t.Parallel()

	handler, ctx := newPaymentProofTestContext(t, "receipt.jpg", []byte("not really an image"))

	if _, err := handler.savePaymentProof(ctx, "payment_proof"); err == nil {
		t.Fatal("expected mismatched file content to be rejected")
	}
}

func TestSavePaymentProofRejectsOversizedFile(t *testing.T) {
	t.Parallel()

	oversized := bytes.Repeat([]byte("a"), maxPaymentProofFileSize+1)
	handler, ctx := newPaymentProofTestContext(t, "receipt.pdf", oversized)

	if _, err := handler.savePaymentProof(ctx, "payment_proof"); err == nil {
		t.Fatal("expected oversized proof upload to be rejected")
	}
}

func TestPaymentProofFilePathRejectsTraversal(t *testing.T) {
	t.Parallel()

	uploadDir := t.TempDir()

	if _, ok := paymentProofFilePath(uploadDir, "/static/payment-proofs/../../etc/passwd"); ok {
		t.Fatal("expected traversal path to be rejected")
	}

	validPath, ok := paymentProofFilePath(uploadDir, "/static/payment-proofs/receipt.pdf")
	if !ok {
		t.Fatal("expected valid payment proof path to be accepted")
	}
	if !strings.HasPrefix(validPath, uploadDir) {
		t.Fatalf("expected valid path to stay inside upload dir %q, got %q", uploadDir, validPath)
	}
}

func newPaymentProofTestContext(t *testing.T, filename string, content []byte) (*OrderHandler, *gin.Context) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("payment_proof", filename)
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err := fileWriter.Write(content); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/order/123/payment-proof", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = req

	return &OrderHandler{paymentProofDir: t.TempDir()}, ctx
}
