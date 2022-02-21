package alidmslim

import (
	"context"
	"os"
	"testing"
)

func TestSend(t *testing.T) {
	accountName, keyId, keySecret := os.Getenv("ALIDM_ACCOUNT"), os.Getenv("ALIDM_KEY_ID"), os.Getenv("ALIDM_KEY_SECRET")
	if accountName == "" || keyId == "" || keySecret == "" {
		t.Fatal("Please set ALIDM_ACCOUNT, ALIDM_KEY_ID and ALIDM_KEY_SECRET environment variables.")
	}

	targetAddress := os.Getenv("TARGET_ADDRESS")
	if targetAddress == "" {
		t.Fatal("Please set TARGET_ADDRESS environment variable.")
	}

	client := NewClient(accountName, keyId, keySecret).Debug(os.Getenv("DEBUG") == "1")
	mail := client.NewHTMLMail("test", "<b>hello</b> world")
	ctx := context.Background()
	mail.MustSend(ctx, targetAddress)
}
