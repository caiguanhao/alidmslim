# alidmslim

Usage:

```go
import "github.com/caiguanhao/alidmslim"

client := alidmslim.NewClient("noreply@yourdomain.com", ACCESS_KEY_ID, ACCESS_KEY_SECRET)

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

client.NewMail(subject, content).Send(ctx, "a@a.com", "b@b.com", ...)
```
