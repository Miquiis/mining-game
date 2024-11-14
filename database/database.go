package database

import (
	"encoding/json"
	"log"

	"github.com/labstack/echo/v4"
	"go.etcd.io/bbolt"
)

var db *bbolt.DB

type User struct {
	Username  string    `json:"username"`
	IP        string    `json:"ip"`
	Gold      int       `json:"gold"`
	Diamond   int       `json:"diamond"`
	Health    int       `json:"health"`
	Inventory Inventory `json:"inventory"`
	Upgrades  []Upgrade `json:"upgrades"`
}

func (u User) Exists() bool {
	return u.Username != ""
}

func (u User) IsLoggedIn() bool {
	return u.IP != ""
}

func (u User) HasPickaxe(pickaxe Pickaxe) bool {
	return u.Inventory.Pickaxe.Id == pickaxe.Id
}

func (u User) HasUpgrade(upgrade Upgrade) bool {
	for _, u := range u.Upgrades {
		if u.Id == upgrade.Id {
			return true
		}
	}
	return false
}

func (u User) HasUpgradeId(upgradeId string) bool {
	for _, u := range u.Upgrades {
		if u.Id == upgradeId {
			return true
		}
	}
	return false
}

func (u User) GetGoldMultiplier() int {
	if u.HasUpgradeId("gold_multiplier_3") {
		return 3
	} else if u.HasUpgradeId("gold_multiplier_2") {
		return 2
	}
	return 1
}

func (u User) GetDiamondMultiplier() int {
	if u.HasUpgradeId("diamond_multiplier_2") {
		return 2
	}
	return 1
}

func (u User) GetDiamondChance() float32 {
	if u.HasUpgradeId("diamond_chance") {
		return 0.05
	}
	return 0.01
}

type Inventory struct {
	Pickaxe Pickaxe `json:"pickaxe"`
}

type Pickaxe struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Price   int    `json:"price"`
	MinGold int    `json:"minGold"`
	MaxGold int    `json:"maxGold"`
}

type Upgrade struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
}

func Start() {
	var err error
	db, err = bbolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func Close() {
	if db != nil {
		db.Close()
	}
}

func GetDB() *bbolt.DB {
	return db
}

func GetUserDataByContext(c echo.Context) User {
	ip := c.RealIP()
	return GetUserDataByIP(ip)
}

func GetUserData(user string) User {
	var value User
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		if b == nil {
			return nil
		}
		data := b.Get([]byte(user))
		if data == nil {
			return nil
		}
		json.Unmarshal(data, &value)
		return nil
	})
	return value
}

func GetUserDataByIP(ip string) User {
	var value User
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		if b == nil {
			return nil
		}

		err := b.ForEach(func(k, v []byte) error {
			var user User
			json.Unmarshal(v, &user)
			if user.IP == ip {
				value = user
			}
			return nil
		})

		if err != nil {
			return err
		}

		return nil
	})
	return value
}

func Login(user string, ip string) {
	db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return err
		}

		userData := GetUserData(user)

		if userData.Username == "" {
			userData.Username = user
			userData.IP = ip
			userData.Inventory.Pickaxe = Pickaxe{
				Name:    "Rusty Pickaxe",
				Price:   0,
				MinGold: 1,
				MaxGold: 5,
			}
			data, err := json.Marshal(userData)
			if err != nil {
				return err
			}
			return b.Put([]byte(user), data)
		}

		userData.IP = ip

		updatedData, err := json.Marshal(userData)
		if err != nil {
			return err
		}

		return b.Put([]byte(user), updatedData)
	})
}

func LogOutUser(user string) {
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		if b == nil {
			return nil
		}

		userData := GetUserData(user)

		userData.IP = ""

		updatedData, err := json.Marshal(userData)
		if err != nil {
			return err
		}

		return b.Put([]byte(user), updatedData)
	})
}

func Mine(userData User, gold int, diamond int) {
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		if b == nil {
			return nil
		}

		userData.Gold += gold
		userData.Diamond += diamond

		updatedData, err := json.Marshal(userData)
		if err != nil {
			return err
		}

		return b.Put([]byte(userData.Username), updatedData)
	})
}

func GetTopScores() []User {
	var scores []User
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		if b == nil {
			return nil
		}

		err := b.ForEach(func(k, v []byte) error {
			var user User
			json.Unmarshal(v, &user)
			if user.Gold == 0 {
				return nil
			}
			scores = append(scores, user)
			return nil
		})

		if err != nil {
			return err
		}

		return nil
	})
	return scores
}

func SaveUserData(userData User) {
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		if b == nil {
			return nil
		}

		data, err := json.Marshal(userData)
		if err != nil {
			return err
		}

		return b.Put([]byte(userData.Username), data)
	})
}
