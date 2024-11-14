package shop

import (
	"encoding/json"
	"fmt"
	"log"
	"mining-game/database"
	"mining-game/htmlwrapper"
	"os"
	"sort"

	"github.com/labstack/echo/v4"
)

type Pickaxe = database.Pickaxe
type Upgrade = database.Upgrade

type ShopItems struct {
	Pickaxes map[string]Pickaxe `json:"pickaxes"`
	Upgrades map[string]Upgrade `json:"upgrades"`
}

type SortedPickaxes []Pickaxe

func (a SortedPickaxes) Len() int           { return len(a) }
func (a SortedPickaxes) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortedPickaxes) Less(i, j int) bool { return a[i].Price < a[j].Price }

type SortedUpgrades []Upgrade

func (a SortedUpgrades) Len() int           { return len(a) }
func (a SortedUpgrades) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortedUpgrades) Less(i, j int) bool { return a[i].Price < a[j].Price }

var pageTitle string = "Mining Game - Shop"

var shopMessage string = `Welcome to the shop!
Here you can buy items to help you mine more gold.

Your current pickaxe: %s.

You have %s gold available.
%s
Buy items using '/shop/<item number>'

Pickaxes:
%s

%s
`

var cachedShopItems ShopItems

var cachedSortedPickaxes SortedPickaxes
var cachedSortedUpgrades SortedUpgrades

func Init() {
	data, err := os.ReadFile("./shop/shop.json")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(data, &cachedShopItems)
	if err != nil {
		log.Fatal(err)
	}

	var pickaxeSlice []Pickaxe
	for _, pickaxe := range cachedShopItems.Pickaxes {
		pickaxeSlice = append(pickaxeSlice, pickaxe)
	}

	sort.Sort(SortedPickaxes(pickaxeSlice))

	cachedSortedPickaxes = pickaxeSlice

	var upgradeSlice []Upgrade

	for _, upgrade := range cachedShopItems.Upgrades {
		upgradeSlice = append(upgradeSlice, upgrade)
	}

	sort.Sort(SortedUpgrades(upgradeSlice))

	cachedSortedUpgrades = upgradeSlice
}

func Shop(c echo.Context) error {
	// Get the user's data
	userData := database.GetUserDataByContext(c)

	if !userData.Exists() {
		return htmlwrapper.PageRedirectWrapper(c, pageTitle, "You need to be logged in to check out the shop!", "/", 3000)
	}

	userGold := fmt.Sprintf("%d", userData.Gold)

	diamondMessage := "You also have %d diamond(s) available.\n"
	upgradesMessage := "Upgrades:\n%s"

	if userData.Diamond > 0 {
		diamondMessage = fmt.Sprintf(diamondMessage, userData.Diamond)
		upgradesMessage = fmt.Sprintf(upgradesMessage, listUpgrades(userData))
	} else {
		diamondMessage = ""
		upgradesMessage = ""
	}

	return htmlwrapper.PageWrapper(c, pageTitle, fmt.Sprintf(shopMessage, userData.Inventory.Pickaxe.Name, userGold, diamondMessage, listPickaxes(userData), upgradesMessage))
}

// GET /shop/:itemId
func ShopItem(c echo.Context) error {
	// Get the user's data
	userData := database.GetUserDataByContext(c)

	if !userData.Exists() {
		return htmlwrapper.PageRedirectWrapper(c, pageTitle, "You need to be logged in to buy an item!", "/", 3000)
	}

	// Get the item number
	itemNumber := c.Param("itemId")

	// Convert the item number to an integer
	itemIndex := 0
	fmt.Sscanf(itemNumber, "%d", &itemIndex)

	// Check if the item number is valid
	if itemIndex < 1 || itemIndex > len(cachedSortedPickaxes) {
		if itemIndex < 1 || itemIndex > len(cachedSortedPickaxes)+len(cachedSortedUpgrades) {
			return htmlwrapper.PageRedirectWrapper(c, pageTitle, "Invalid item number!", "/shop", 3000)
		}
		return handleUpgradePurchase(c, userData, itemIndex)
	}

	// Get the pickaxe
	pickaxe := cachedSortedPickaxes[itemIndex-1]

	// Check if the user has enough gold
	if userData.Gold < pickaxe.Price {
		return htmlwrapper.PageRedirectWrapper(c, pageTitle, "You don't have enough gold to buy that item!", "/shop", 3000)
	}

	// Update the user's data
	userData.Gold -= pickaxe.Price
	userData.Inventory.Pickaxe = pickaxe

	// Save the user's data
	database.SaveUserData(userData)

	// Redirect the user to the shop
	return htmlwrapper.PageRedirectWrapper(c, pageTitle, fmt.Sprintf("You have successfully bought %s!", pickaxe.Name), "/shop", 3000)
}

func handleUpgradePurchase(c echo.Context, userData database.User, itemIndex int) error {
	// Get the upgrade
	upgrade := cachedSortedUpgrades[itemIndex-1-len(cachedSortedPickaxes)]

	// Check if the user has enough gold
	if userData.Diamond < upgrade.Price {
		return htmlwrapper.PageRedirectWrapper(c, pageTitle, "You don't have enough diamonds to buy that item!", "/shop", 3000)
	}

	// Update the user's data
	userData.Diamond -= upgrade.Price
	userData.Upgrades = append(userData.Upgrades, upgrade)

	// Save the user's data
	database.SaveUserData(userData)

	// Redirect the user to the shop
	return htmlwrapper.PageRedirectWrapper(c, pageTitle, fmt.Sprintf("You have successfully bought %s!", upgrade.Name), "/shop", 3000)
}

func listPickaxes(user database.User) string {
	var message string
	for index, pickaxe := range cachedSortedPickaxes {
		if user.HasPickaxe(pickaxe) {
			message += fmt.Sprintf("[%d] - %s: %d gold | %d-%d range | ALREADY OWN\n", index+1, pickaxe.Name, pickaxe.Price, pickaxe.MinGold, pickaxe.MaxGold)
		} else {
			message += fmt.Sprintf("[%d] - %s: %d gold | %d-%d range\n", index+1, pickaxe.Name, pickaxe.Price, pickaxe.MinGold, pickaxe.MaxGold)
		}
	}
	return message
}

func listUpgrades(user database.User) string {
	var message string
	pickaxesLen := len(cachedSortedPickaxes)
	for index, upgrade := range cachedSortedUpgrades {
		if user.HasUpgrade(upgrade) {
			message += fmt.Sprintf("[%d] - %s: %d diamond | ALREADY OWN\n  • %s\n", index+1+pickaxesLen, upgrade.Name, upgrade.Price, upgrade.Description)
		} else {
			message += fmt.Sprintf("[%d] - %s: %d diamond\n  • %s\n", index+1+pickaxesLen, upgrade.Name, upgrade.Price, upgrade.Description)
		}
	}
	return message
}
