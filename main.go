package main

import (
	"fmt"
	"math/rand"
	"slices"
	"time"
)

// type repeating date
type user struct {
	id      int
	friends []int
}
type schedule struct {
	st    time.Time
	end   time.Time
	inter time.Duration
}
type meeting struct {
	u1, u2 int //more users per meeting - maybe cancel cross communication
	s      time.Time
}
type database struct {
	users    map[int]*user
	meetings map[int]map[int][]schedule // (uid1, uid2) > []schedule
}
type route struct {
	currentlyVisiting int
	meetings          []meeting
	nodes             []int
	earliestTime      time.Time
}

func NewDB() *database {
	return &database{
		users:    make(map[int]*user),
		meetings: make(map[int]map[int][]schedule),
	}
}

func main() {
	k := 1000
	conn := 10
	searchRoutes := 10
	maxHops := 5
	timeout := time.Hour * 5

	agg := 0
	db := NewDB()

	for i := 0; i < k; i++ {
		db.addUser(i)
	}

	for _, u := range db.users {
		for i := 0; i < conn; i++ {
			fr := rand.Intn(k)
			db.addFriend(u, fr)
			db.addSchedule(u, fr, time.Now().Add(time.Duration(rand.Intn(10))*time.Hour), time.Hour, time.Hour*time.Duration(rand.Intn(30)))
		}
	}
	for i := 0; i < searchRoutes; i++ {
		target := rand.Intn(k)
		rs, _ := db.findRoutes(db.users[i], target, timeout, maxHops)
		agg += len(rs)
	}

	fmt.Println("%routes found", float64(agg)/float64(searchRoutes))
}

func (db *database) findRoutes(startU *user, goalId int, timeout time.Duration, maxMeetings int) ([]route, error) {
	routes := make([]route, 0)
	toVisit := []route{{currentlyVisiting: startU.id, nodes: []int{startU.id}, earliestTime: time.Now()}}

	hops := 0
	for len(toVisit) > 0 {
		hops++
		if hops%10000 == 0 {
			fmt.Printf("hop %v | currently in queue: %v\n", hops, len(toVisit))
		}

		r := toVisit[0]
		toVisit = toVisit[1:]

		u, ok := db.users[r.currentlyVisiting]
		if !ok {
			panic("not found")
		}

		for _, fr := range u.friends {
			if slices.Contains(r.nodes, fr) {
				continue
			}

			a, b := u.id, fr
			if a > b {
				a, b = b, a
			}
			schedsA, exA := db.meetings[a]
			if !exA {
				continue
			}
			scheds, exB := schedsA[b]
			if !exB {
				continue
			}

			earliest := time.Now().Add(timeout).Add(time.Hour)
			earliestDup := earliest
			for _, sched := range scheds {
				if sched.inter == 0 {
					if sched.end.After(r.earliestTime) {
						earliestMeet := sched.st
						if sched.st.Before(r.earliestTime) {
							earliestMeet = r.earliestTime
						}
						if earliestMeet.Before(earliest) {
							earliest = earliestMeet
						}
					}
					continue
				}

				if sched.end.After(r.earliestTime) || r.earliestTime.Equal(sched.end) {
					earliestMeet := sched.st
					if sched.st.Before(r.earliestTime) {
						earliestMeet = r.earliestTime
					}
					if earliestMeet.Before(earliest) {
						earliest = earliestMeet
					}
					continue
				}

				diffT := r.earliestTime.Sub(sched.st)
				cnts := diffT / sched.inter

				t1st := sched.end.Add(sched.inter * cnts)
				t1end := sched.end.Add(sched.inter * cnts)

				if t1end.Before(r.earliestTime) {
					t1st = t1st.Add(sched.inter)
					t1end = t1end.Add(sched.inter)
				}

				if (t1st.Before(r.earliestTime) || t1st.Equal(r.earliestTime)) &&
					(t1end.After(r.earliestTime) || t1end.Equal(r.earliestTime)) {
					earliestMeet := t1st
					if t1st.Before(r.earliestTime) {
						earliestMeet = r.earliestTime
					}
					if earliestMeet.Before(earliest) {
						earliest = earliestMeet
					}
				}
			}

			if earliest.Equal(earliestDup) {
				continue
			}

			if time.Until(earliest) > timeout {
				continue
			}

			rNew := route{
				currentlyVisiting: fr,
				nodes:             append(r.nodes, fr),
				meetings:          append(r.meetings, meeting{u1: a, u2: b, s: earliest}),
				earliestTime:      earliest,
			}

			if fr == goalId {
				routes = append(routes, rNew)
			} else if len(rNew.meetings) <= maxMeetings {
				toVisit = append(toVisit, rNew)
			}
		}

	}

	return routes, nil
}

func (db *database) addSchedule(u1 *user, u2Id int, date time.Time, dur time.Duration, interval time.Duration) error {

	if !slices.Contains(u1.friends, u2Id) {
		return fmt.Errorf("You can only add friends, %v is not in your friendlist yet", u2Id)
	}

	u2, ok := db.users[u2Id]
	if !ok {
		return fmt.Errorf("User %v not found, unable to add schedule", u2Id)
	}

	ua, ub := u1.id, u2.id
	if ua > ub {
		ua, ub = ub, ua
	}

	var scheds []schedule
	schedsA, exA := db.meetings[ua]
	if exA {
		scheds, _ = schedsA[ub]
	}

	// search if schedule already exists - else add it
	// todo: merge with existing schedules
	for _, sc := range scheds {
		if sc.st.Equal(date) && sc.inter == interval {
			return fmt.Errorf("schedule already exists")
		}
	}

	if !exA {
		db.meetings[ua] = make(map[int][]schedule)
	}

	if scheds == nil {
		db.meetings[ua][ub] = make([]schedule, 0)
	}

	scheds = append(scheds, schedule{st: date, inter: interval, end: date.Add(dur)})
	db.meetings[ua][ub] = scheds

	return nil
}

func (db *database) removeFriend(u1 *user, u2Id int) error {
	if u1.id == u2Id {
		return fmt.Errorf("you cannot unfriend yourself :(")
	}

	foundAtLeastOne := false
	u2Ind := slices.Index(u1.friends, u2Id)
	if u2Ind != -1 {
		foundAtLeastOne = true
		u1.friends[u2Ind] = u1.friends[len(u1.friends)-1]
		u1.friends = u1.friends[:len(u1.friends)-1]
	}

	if u2, u2Ex := db.users[u2Id]; u2Ex {
		foundAtLeastOne = true
		u1Ind := slices.Index(u2.friends, u1.id)
		if u1Ind != -1 {
			u2.friends[u1Ind] = u2.friends[len(u2.friends)-1]
			u2.friends = u2.friends[:len(u2.friends)-1]
		}
	}

	if !foundAtLeastOne {
		return fmt.Errorf("id %v was not in your friendlist, cannot remove", u2Id)
	}
	return nil
}

func (db *database) addFriend(u1 *user, u2Id int) error {
	if u1.id == u2Id {
		return fmt.Errorf("you cannot friend yourself :(")
	}
	u2, ex := db.users[u2Id]
	if !ex {
		return fmt.Errorf("user %v does not exist, could not add friend", u2Id)
	}

	if !slices.Contains(u1.friends, u2Id) {
		u1.friends = append(u1.friends, u2.id)
	}

	if !slices.Contains(u2.friends, u1.id) {
		u2.friends = append(u2.friends, u1.id)
	}

	return nil
}

func (db *database) addUser(id int) (*user, error) {
	if _, ok := db.users[id]; ok {
		return nil, fmt.Errorf("user %v already exists", id)
	}

	u := &user{id: id}
	db.users[id] = u
	return u, nil
}
