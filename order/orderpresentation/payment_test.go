package orderpresentation

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSavePaymentProofFailsWhenCloudinaryNotConfigured(t *testing.T) {
	t.Parallel()

	handler, ctx := newPaymentProofTestContext(t, "receipt.pdf", []byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj\n"))

	// The test handler has no Cloudinary credentials wired in, so a
	// well-formed upload should pass file validation and then fail cleanly
	// at the "where do we send it" step, rather than attempting a real
	// network call to Cloudinary.
	if _, err := handler.savePaymentProof(ctx, "payment_proof", 123); !errors.Is(err, errPaymentProofStorageUnavailable) {
		t.Fatalf("expected errPaymentProofStorageUnavailable, got %v", err)
	}
}

func TestSavePaymentProofRejectsMismatchedContentType(t *testing.T) {
	t.Parallel()

	handler, ctx := newPaymentProofTestContext(t, "receipt.jpg", []byte("not really an image"))

	if _, err := handler.savePaymentProof(ctx, "payment_proof", 123); err == nil {
		t.Fatal("expected mismatched file content to be rejected")
	}
}

func TestSavePaymentProofRejectsOversizedFile(t *testing.T) {
	t.Parallel()

	oversized := bytes.Repeat([]byte("a"), maxPaymentProofFileSize+1)
	handler, ctx := newPaymentProofTestContext(t, "receipt.pdf", oversized)

	if _, err := handler.savePaymentProof(ctx, "payment_proof", 123); err == nil {
		t.Fatal("expected oversized proof upload to be rejected")
	}
}

func TestCloudinaryProofUploaderConfigured(t *testing.T) {
	t.Parallel()

	var nilUploader *cloudinaryProofUploader
	if nilUploader.configured() {
		t.Fatal("expected nil uploader to report not configured")
	}

	if newCloudinaryProofUploader("", "key", "secret").configured() {
		t.Fatal("expected uploader with missing cloud name to report not configured")
	}
	if !newCloudinaryProofUploader("cloud", "key", "secret").configured() {
		t.Fatal("expected uploader with all credentials to report configured")
	}
}

func TestIsCloudinaryProofURL(t *testing.T) {
	t.Parallel()

	if !isCloudinaryProofURL("https://res.cloudinary.com/demo/image/upload/v1/payment-proofs/1/abc.jpg") {
		t.Fatal("expected a res.cloudinary.com URL to be recognized")
	}
	if isCloudinaryProofURL("some-local-filename.jpg") {
		t.Fatal("expected a bare legacy filename to not be recognized as a Cloudinary URL")
	}
	if isCloudinaryProofURL("https://evil.example.com/res.cloudinary.com/x.jpg") {
		t.Fatal("expected a lookalike host to be rejected")
	}
}

func TestPaymentProofFilePathRejectsTraversal(t *testing.T) {
	t.Parallel()

	uploadDir := t.TempDir()

	if _, ok := paymentProofFilePath(uploadDir, "../../etc/passwd"); ok {
		t.Fatal("expected traversal path to be rejected")
	}
	if _, ok := paymentProofFilePath(uploadDir, "sub/receipt.pdf"); ok {
		t.Fatal("expected non-bare filename to be rejected")
	}

	validPath, ok := paymentProofFilePath(uploadDir, "receipt.pdf")
	if !ok {
		t.Fatal("expected valid payment proof filename to be accepted")
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

	return &OrderHandler{paymentProofDir: t.TempDir(), proofUploader: newCloudinaryProofUploader("", "", "")}, ctx
}
