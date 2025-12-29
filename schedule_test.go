package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestFriends(t *testing.T) {
	db := NewDB()
	var err error

	u0, _ := db.addUser(0)
	u1, _ := db.addUser(1)
	u2, _ := db.addUser(2)

	err = db.addFriend(u0, u1.id)
	if err != nil {
		t.Error(err)
	}
	err = db.addFriend(u0, u2.id)
	if err != nil {
		t.Error(err)
	}

	if len(u0.friends) != 2 {
		t.Errorf("u0: not 2 friends but %v", len(u0.friends))
	}

	if len(u1.friends) != 1 {
		t.Errorf("u1: not 1 friends but %v", len(u1.friends))
	}

	err = db.removeFriend(u0, u1.id)
	if err != nil {
		t.Error(err)
	}

	if len(u0.friends) != 1 {
		t.Errorf("u0: not 1 friends but %v", len(u1.friends))
	}
	if len(u1.friends) != 0 {
		t.Errorf("u1: not 0 friends but %v", len(u1.friends))
	}
}

func TestSimpleRouting(t *testing.T) {
	db := NewDB()
	names := map[string]int{
		"bennus":   0,
		"herrmann": 1,
		"fga":      2,
		"stefan":   3,
		"luke":     4,
	}

	for _, uid := range names {
		db.addUser(uid)
	}

	for _, uid := range names {
		for _, uid2 := range names {
			db.addFriend(db.users[uid], uid2)
		}
	}

	monday := time.Date(2025, 12, 30, 21, 0, 0, 0, time.Local)
	db.addSchedule(db.users[names["bennus"]], names["fga"], monday, time.Hour*2, time.Hour*24*7)
	db.addSchedule(db.users[names["stefan"]], names["fga"], monday, time.Hour*2, time.Hour*24*7)
	db.addSchedule(db.users[names["luke"]], names["fga"], monday, time.Hour*2, time.Hour*24*7)
	db.addSchedule(db.users[names["herrmann"]], names["bennus"], time.Date(2025, 12, 29, 21, 0, 0, 0, time.Local), time.Hour, time.Hour*24*7)
	db.addSchedule(db.users[names["herrmann"]], names["luke"], time.Date(2025, 12, 29, 18, 0, 0, 0, time.Local), time.Hour, time.Hour*24*7)

	routes, _ := db.findRoutes(monday.AddDate(0, 0, -2), db.users[names["herrmann"]], names["stefan"], time.Hour*24*15, 5)

	if len(routes) != 2 {
		t.Errorf("should be 2 routes, found %v", len(routes))
	}

	for _, r := range routes {
		if len(r.meetings) != 3 {
			t.Errorf("routes should be length 3, found %v", len(r.meetings))
		}
	}

}

// 644554928ns 6 hops 1 route
func BenchmarkRouting(b *testing.B) {
	rng := rand.New(rand.NewSource(404))
	k := 10000
	conn := 10
	searchRoutes := 1
	maxHops := 6
	timeout := time.Hour * 5

	agg := 0
	for n := 0; n < b.N; n++ {
		db := NewDB()

		for i := 0; i < k; i++ {
			db.addUser(i)
		}

		for j := 0; j < k; j++ {
			u := db.users[j]
			for i := 0; i < conn; i++ {
				fr := rng.Intn(k)
				db.addFriend(u, fr)
				db.addSchedule(u, fr, time.Now().Add(time.Duration(rng.Intn(10))*time.Hour), time.Hour, time.Hour*time.Duration(rng.Intn(30)))
			}
		}
		for i := 0; i < searchRoutes; i++ {
			target := rng.Intn(k)
			rs, _ := db.findRoutes(time.Now(), db.users[i], target, timeout, maxHops)
			agg += len(rs)
		}
	}
	fmt.Println(float64(agg) / float64(b.N) / float64(searchRoutes))
}
