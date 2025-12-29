package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"slices"
	"strings"
	"time"
)

// type repeating date
type user struct {
	id       int
	friends  []int
	secret   string
	token    string
	tokenEnd time.Time
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
	bad               bool
}

func NewDB() *database {
	return &database{
		users:    make(map[int]*user),
		meetings: make(map[int]map[int][]schedule),
	}
}

func main() {
	db := NewDB()
	launchServer(db)
}

func launchServer(db *database) {

	http.HandleFunc("/b/login", db.authHandler)
	http.HandleFunc("/b/register", db.registerHandler)
	http.HandleFunc("/b/addFriend", db.addFriendHandler)

	http.ListenAndServe("localhost:8080", http.DefaultServeMux)
}

func (db *database) addFriendHandler(w http.ResponseWriter, r *http.Request) {

}

func (db *database) IsLoggedIn(r *http.Request) int {
	reqToken := r.Header.Get("Authorization")
	reqToken = strings.TrimPrefix(reqToken, "Bearer ")

	uid := -1
	for _, u := range db.users {
		if u.token == reqToken && u.tokenEnd.After(time.Now()) {
			return u.id
		}
	}
	return uid
}

func (db *database) authHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	pw := fmt.Sprintf("%v", time.Now())
	uid := -1

	decoder := json.NewDecoder(r.Body)
	var creds struct {
		Uid int
		Pw  string
	}
	err := decoder.Decode(&creds)
	if err != nil {
		uid = creds.Uid
		pw = creds.Pw
	}
	userPwHash := ""
	u := &user{}
	if uid != -1 {
		if u2, ok := db.users[uid]; ok {
			u = u2
		}
	}
	userPwHash = u.secret

	pwHash := sha256.Sum256([]byte(pw))

	if subtle.ConstantTimeCompare(pwHash[:], []byte(userPwHash)[:]) == 1 && u.id == uid {
		token := sha256.Sum256([]byte(fmt.Sprintf("%v%v", uid, time.Now())))
		u.token = string(token[:])
		u.tokenEnd = time.Now().Add(time.Hour * 4)

		enc := json.NewEncoder(w)
		enc.Encode(struct{ token string }{u.token})
		w.WriteHeader(200)
		fmt.Printf("logged in user %v", uid)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func (db *database) registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var newPW struct {
		Pw string `json:"pw"`
	}
	err := decoder.Decode(&newPW)
	if err == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	uid := -1
	for {
		uid = rand.Intn(1000)
		u, err := db.addUser(uid)
		if err == nil {
			sec := sha256.Sum256([]byte(newPW.Pw))
			u.secret = string(sec[:])
			break
		}
	}

	if uid != -1 {
		fmt.Printf("added new user %v", uid)
		wr := json.NewEncoder(w)
		wr.Encode(struct{ Uid int }{Uid: uid})
		w.WriteHeader(200)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (db *database) findRoutes(startTime time.Time, startU *user, goalId int, timeout time.Duration, maxMeetings int) ([]route, error) {
	routes := make([]route, 0)
	toVisit := []route{{currentlyVisiting: startU.id, nodes: []int{startU.id}, earliestTime: startTime}}
	toVisitNext := []route{}

	hops := 0
	depth := 0
	for len(toVisit) > 0 || (len(toVisitNext) > 0 && depth < maxMeetings+1) {
		hops++
		if hops%10000 == 0 {
			fmt.Printf("hop %v | currently in queue: %v\n", hops, len(toVisit))
		}

		if len(toVisit) == 0 {
			depth++
			fmt.Printf("Reached depth %v\n", depth)
			toVisit = toVisitNext
			toVisitNext = []route{}
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

			earliest := startTime.Add(timeout).Add(time.Hour)
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

			if earliest.Sub(startTime) > timeout {
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
				toVisitNext = append(toVisitNext, rNew)
			}
		}

	}

	return routes, nil
}

func (db *database) findRoutesMulti(startTime time.Time, startU *user, goalId int, timeout time.Duration, maxMeetings int) ([]route, error) {
	workerCount := 64
	routes := make([]route, 0)
	toVisit := []route{{currentlyVisiting: startU.id, nodes: []int{startU.id}, earliestTime: startTime}}
	toVisitNext := []route{}

	depth := 0
	for depth < maxMeetings+1 {
		work := make(chan route, 1)
		go func() {
			for _, r := range toVisit {
				work <- r
			}
		}()

		type result struct {
			toV  []route
			goal []route
		}
		results := make(chan result, 1)

		worker := func() {
			for {
				r2, more := <-work
				if !more {
					return
				}
				res := result{}
				u, ok := db.users[r2.currentlyVisiting]
				if !ok {
					results <- res
					return
				}

				for _, fr := range u.friends {
					if slices.Contains(r2.nodes, fr) {
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

					earliest := startTime.Add(timeout).Add(time.Hour)
					earliestDup := earliest
					for _, sched := range scheds {
						if sched.inter == 0 {
							if sched.end.After(r2.earliestTime) {
								earliestMeet := sched.st
								if sched.st.Before(r2.earliestTime) {
									earliestMeet = r2.earliestTime
								}
								if earliestMeet.Before(earliest) {
									earliest = earliestMeet
								}
							}
							continue
						}

						if sched.end.After(r2.earliestTime) || r2.earliestTime.Equal(sched.end) {
							earliestMeet := sched.st
							if sched.st.Before(r2.earliestTime) {
								earliestMeet = r2.earliestTime
							}
							if earliestMeet.Before(earliest) {
								earliest = earliestMeet
							}
							continue
						}

						diffT := r2.earliestTime.Sub(sched.st)
						cnts := diffT / sched.inter

						t1st := sched.end.Add(sched.inter * cnts)
						t1end := sched.end.Add(sched.inter * cnts)

						if t1end.Before(r2.earliestTime) {
							t1st = t1st.Add(sched.inter)
							t1end = t1end.Add(sched.inter)
						}

						if (t1st.Before(r2.earliestTime) || t1st.Equal(r2.earliestTime)) &&
							(t1end.After(r2.earliestTime) || t1end.Equal(r2.earliestTime)) {
							earliestMeet := t1st
							if t1st.Before(r2.earliestTime) {
								earliestMeet = r2.earliestTime
							}
							if earliestMeet.Before(earliest) {
								earliest = earliestMeet
							}
						}
					}

					if earliest.Equal(earliestDup) {
						continue
					}

					if earliest.Sub(startTime) > timeout {
						continue
					}

					rNew := route{
						currentlyVisiting: fr,
						nodes:             append(r2.nodes, fr),
						meetings:          append(r2.meetings, meeting{u1: a, u2: b, s: earliest}),
						earliestTime:      earliest,
					}

					if fr == goalId {
						res.goal = append(res.goal, rNew)
					} else if len(rNew.meetings) <= maxMeetings {
						res.toV = append(res.toV, rNew)
					}
				}
				results <- res
			}
		}

		for i := 0; i < workerCount; i++ {
			go worker()
		}

		for i := 0; i < len(toVisit); i++ {
			res := <-results
			if len(res.toV) > 0 {
				toVisitNext = append(toVisitNext, res.toV...)
			}
			if len(res.goal) > 0 {
				routes = append(routes, res.toV...)
			}
		}

		depth++
		fmt.Printf("Reached depth %v\n", depth)
		toVisit = toVisitNext
		toVisitNext = []route{}

		close(work)
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
