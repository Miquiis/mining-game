package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sort"

	"github.com/labstack/echo/v4"

	"mining-game/database"
	"mining-game/htmlwrapper"
	"mining-game/shop"
)

type ByGold []database.User

func (a ByGold) Len() int           { return len(a) }
func (a ByGold) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByGold) Less(i, j int) bool { return a[i].Gold > a[j].Gold }

var pageName string = "Mining Game - Home"

var homeMessage string = `Welcome back, %s!
%s

Start out by going to /mine
You can also visit the shop at /shop
Access the scoreboard at /scoreboard

Logout using /logout
`

var goldMessage string = `You have %s gold.`
var diamondMessage string = `You have %s gold and %s diamond(s).`

func main() {
	database.Start()
	shop.Init()

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		ip := c.RealIP()
		userData := database.GetUserDataByIP(ip)
		if userData.Username != "" {
			gold := fmt.Sprintf("%d", userData.Gold)
			diamond := fmt.Sprintf("%d", userData.Diamond)

			currencyMessage := fmt.Sprintf(goldMessage, gold)
			if userData.Diamond > 0 {
				currencyMessage = fmt.Sprintf(diamondMessage, gold, diamond)
			}

			return htmlwrapper.PageWrapper(c, pageName, fmt.Sprintf(homeMessage, userData.Username, currencyMessage))
		}
		return htmlwrapper.PageWrapper(c, pageName, "Hello, there!\nLogin using /login?username=yourname")
	})
	e.GET("/login", login)
	e.GET("/logout", logout)
	e.GET("/mine", mine)
	e.GET("/scoreboard", scoreboard)
	e.GET("/shop", shop.Shop)
	e.GET("/shop/:itemId", shop.ShopItem)
	e.Logger.Fatal(e.Start(":3288"))
}

func scoreboard(c echo.Context) error {
	scores := database.GetTopScores()

	sort.Sort(ByGold(scores))

	var message string
	for i, score := range scores {
		if score.Gold == 0 {
			continue
		}
		message += fmt.Sprintf("%d. %s: %d gold\n", i+1, score.Username, score.Gold)
	}

	return htmlwrapper.PageWrapper(c, "Mining Game - Scoreboard", message)
}

func mine(c echo.Context) error {
	ip := c.RealIP()

	userData := database.GetUserDataByIP(ip)

	if userData.Username == "" {
		return htmlwrapper.PageRedirectWrapper(c, "Mining Game - Mine", "You need to be logged in to mine!", "/", 3000)
	}

	minGold := userData.Inventory.Pickaxe.MinGold
	maxGold := userData.Inventory.Pickaxe.MaxGold

	minedGold := (rand.Intn(maxGold-minGold) + minGold) * userData.GetGoldMultiplier()
	totalGold := fmt.Sprintf("%d", userData.Gold+minedGold)

	message := fmt.Sprintf("You have %s gold in total.", totalGold)
	message += fmt.Sprintf("\nYou mined %d gold!", minedGold)

	minedDiamonds := 0

	if rand.Float32() < userData.GetDiamondChance() {
		minedDiamonds = 1 * userData.GetDiamondMultiplier()
		message += "\nYou found a diamond!"
	}

	database.Mine(userData, minedGold, minedDiamonds)

	if userData.HasUpgradeId("auto_miner") {
		return htmlwrapper.PageRedirectWrapper(c, "Mining Game - Mine", message, "/mine", 3000)
	} else {
		return htmlwrapper.PageWrapper(c, "Mining Game - Mine", message)
	}
}

// func autoRefreshPage(c echo.Context, message string, interval int) error {
// 	return c.HTML(http.StatusOK, fmt.Sprintf(`
// 		<html>
// 		<head>
// 			<meta name="color-scheme" content="light dark">
// 		</head>
// 		<body>
// 			<pre style="word-wrap: break-word; white-space: pre-wrap;" id="dataContainer">%s</pre>
// 			<script>
// 				setTimeout(() => {
// 					location.reload();
// 				}, %d);
// 			</script>
// 		</body>
// 		</html>
// 	`, message, interval))
// }

func logout(c echo.Context) error {
	ip := c.RealIP()

	userData := database.GetUserDataByIP(ip)

	if userData.Username == "" {
		return c.Redirect(http.StatusFound, "/")
	}

	database.LogOutUser(userData.Username)

	return htmlwrapper.PageRedirectWrapper(c, "Mining Game - Logout", "Logged out. See you soon!", "/", 3000)
}

func login(c echo.Context) error {
	username := c.QueryParam("username")

	if username == "" {
		return htmlwrapper.PageRedirectWrapper(c, "Mining Game - Login", "No username provided. \nExample: /login?username=admin", "/", 5000)
	}

	database.Login(username, c.RealIP())

	return htmlwrapper.PageRedirectWrapper(c, "Mining Game - Login", "Logged in as "+username, "/", 3000)
}
