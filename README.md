# reX

Reverse Engineered Twitter API

## Login

```golang
uname := os.Getenv("USER_NAME")
upwd := os.Getenv("PASSWORD")
x := reX.New(uname, upwd)
wd, _ := os.Getwd()
cookiesPath := fmt.Sprintf("%s/cookies.json", wd)
err := x.SetCookies(cookiesPath)
if err != nil || !x.IsLoggedIn() {
    println("You must login first")
    x.Login()
    x.SaveCookies(cookiesPath)
}
```

## Followings

```golang
var cursor *string
for {
    tweets, nextCursor := x.GetFollowingsByScreenName("shareverse_", cursor)
    cursor = nextCursor
    if cursor == nil {
        break
    }
    for _, tweet := range tweets {
        println(tweet.ScreenName)
    }
}
```
