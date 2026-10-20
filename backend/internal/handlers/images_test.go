package handlers

import (
	"bytes"
	"image"
	"image/png"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUploadRejectsInvalidPixelsBeforeEncoding(t *testing.T) {
	var wrongSize bytes.Buffer
	if err := png.Encode(&wrongSize, image.NewRGBA(image.Rect(0, 0, 4, 5))); err != nil {
		t.Fatal(err)
	}
	for _, body := range [][]byte{[]byte("not an image"), wrongSize.Bytes()} {
		request := httptest.NewRequest("POST", "/api/admin/activity-images", bytes.NewReader(body))
		response := httptest.NewRecorder()
		(&Handler{}).AdminUploadActivityImage(response, request)
		if response.Code != 422 {
			t.Fatalf("expected invalid crop status, got %d: %s", response.Code, response.Body)
		}
	}
}

func TestActivityImageNameCannotEscapeStorage(t *testing.T) {
	for _, name := range []string{"../secret.avif", "file.svg", strings.Repeat("a", 32) + ".avif/other", "test.avif"} {
		if activityImageName.MatchString(name) {
			t.Fatalf("unsafe image name accepted: %s", name)
		}
	}
}
