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

func BenchmarkRouting(b *testing.B) {

	k := 1000
	conn := 18

	agg := 0
	for n := 0; n < b.N; n++ {
		db := NewDB()

		for i := 0; i < k; i++ {
			db.addUser(i)
		}

		for _, u := range db.users {
			for i := 0; i < conn; i++ {
				db.addFriend(u, rand.Intn(k))
				db.addSchedule(u, i, time.Now().Add(time.Duration(rand.Intn(10))*time.Hour), time.Hour, time.Hour*time.Duration(rand.Intn(30)))
			}
		}
		for i := 0; i < 5; i++ {
			target := rand.Intn(100)
			rs, _ := db.findRoutes(db.users[i], target, time.Hour*10)
			agg += len(rs)
		}
	}
	fmt.Println(float64(agg) / float64(b.N))
}
