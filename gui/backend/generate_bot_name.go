package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/stellar/kelp/support/utils"
)

var idx_name = 0
var idx_adjectives = 0
var idx_animals = 0

func init() {
	utils.Shuffle(names)
	utils.Shuffle(adjectives)
	utils.Shuffle(ocean_animals)
}

type genBotNameRequest struct {
	UserData UserData `json:"user_data"`
}

func (s *APIServer) generateBotName(w http.ResponseWriter, r *http.Request) {
	bodyBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error when reading request input: %s\n", e))
		return
	}
	var req genBotNameRequest
	e = json.Unmarshal(bodyBytes, &req)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error unmarshaling json: %s; bodyString = %s", e, string(bodyBytes)))
		return
	}
	if strings.TrimSpace(req.UserData.ID) == "" {
		s.writeErrorJson(w, fmt.Sprintf("cannot have empty userID"))
		return
	}

	botName, e := s.doGenerateBotName(req.UserData)
	if e != nil {
		s.writeError(w, fmt.Sprintf("error encountered while generating new bot name: %s", e))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(botName))
}

func (s *APIServer) doGenerateBotName(userData UserData) (string, error) {
	var botName string
	startingIdxName := idx_name
	for {
		name := names[idx_name]
		botName = strings.Title(fmt.Sprintf("%s The %s %s", name, adjectives[idx_adjectives], ocean_animals[idx_animals]))
		idx_name = (idx_name + 1) % len(names)
		idx_adjectives = (idx_adjectives + 1) % len(adjectives)
		idx_animals = (idx_animals + 1) % len(ocean_animals)

		// only check name so we use unique first names for convenience
		prefixExists, e := s.prefixExists(userData, strings.ToLower(name))
		if e != nil {
			return "", fmt.Errorf("error encountered while checking for new bot name prefix: %s", e)
		}
		if !prefixExists {
			break
		}

		// if prefix exists and we have reached back to the starting index then we have exhausted all options
		if idx_name == startingIdxName {
			return "", fmt.Errorf("cannot generate name because we ran out of first name combinations")
		}
	}
	return botName, nil
}

func (s *APIServer) prefixExists(userData UserData, prefix string) (bool, error) {
	command := fmt.Sprintf("ls %s | grep %s", s.botConfigsPathForUser(userData.ID).Unix(), prefix)
	_, e := s.kos.Blocking(userData.ID, "prefix", command)
	if e != nil {
		if strings.Contains(e.Error(), "exit status 1") {
			return false, nil
		}
		return false, fmt.Errorf("error checking for prefix '%s': %s", prefix, e)
	}
	return true, nil
}

// https://www.ssa.gov/oact/babynames/decades/century.html
var names = []string{
	"James",
	"Mary",
	"John",
	"Patricia",
	"Robert",
	"Jennifer",
	"Michael",
	"Linda",
	"William",
	"Elizabeth",
	"David",
	"Barbara",
	"Richard",
	"Susan",
	"Joseph",
	"Jessica",
	"Thomas",
	"Sarah",
	"Charles",
	"Karen",
	"Christopher",
	"Nancy",
	"Daniel",
	"Margaret",
	"Matthew",
	"Lisa",
	"Anthony",
	"Betty",
	"Donald",
	"Dorothy",
	"Mark",
	"Sandra",
	"Paul",
	"Ashley",
	"Steven",
	"Kimberly",
	"Andrew",
	"Donna",
	"Kenneth",
	"Emily",
	"Joshua",
	"Michelle",
	"George",
	"Carol",
	"Kevin",
	"Amanda",
	"Brian",
	"Melissa",
	"Edward",
	"Deborah",
	"Ronald",
	"Stephanie",
	"Timothy",
	"Rebecca",
	"Jason",
	"Laura",
	"Jeffrey",
	"Sharon",
	"Ryan",
	"Cynthia",
	"Jacob",
	"Kathleen",
	"Gary",
	"Helen",
	"Nicholas",
	"Amy",
	"Eric",
	"Shirley",
	"Stephen",
	"Angela",
	"Jonathan",
	"Anna",
	"Larry",
	"Brenda",
	"Justin",
	"Pamela",
	"Scott",
	"Nicole",
	"Brandon",
	"Ruth",
	"Frank",
	"Katherine",
	"Benjamin",
	"Samantha",
	"Gregory",
	"Christine",
	"Samuel",
	"Emma",
	"Raymond",
	"Catherine",
	"Patrick",
	"Debra",
	"Alexander",
	"Virginia",
	"Jack",
	"Rachel",
	"Dennis",
	"Carolyn",
	"Jerry",
	"Janet",
	"Tyler",
	"Maria",
	"Aaron",
	"Heather",
	"Jose",
	"Diane",
	"Henry",
	"Julie",
	"Douglas",
	"Joyce",
	"Adam",
	"Victoria",
	"Peter",
	"Kelly",
	"Nathan",
	"Christina",
	"Zachary",
	"Joan",
	"Walter",
	"Evelyn",
	"Kyle",
	"Lauren",
	"Harold",
	"Judith",
	"Carl",
	"Olivia",
	"Jeremy",
	"Frances",
	"Keith",
	"Martha",
	"Roger",
	"Cheryl",
	"Gerald",
	"Megan",
	"Ethan",
	"Andrea",
	"Arthur",
	"Hannah",
	"Terry",
	"Jacqueline",
	"Christian",
	"Ann",
	"Sean",
	"Jean",
	"Lawrence",
	"Alice",
	"Austin",
	"Kathryn",
	"Joe",
	"Gloria",
	"Noah",
	"Teresa",
	"Jesse",
	"Doris",
	"Albert",
	"Sara",
	"Bryan",
	"Janice",
	"Billy",
	"Julia",
	"Bruce",
	"Marie",
	"Willie",
	"Madison",
	"Jordan",
	"Grace",
	"Dylan",
	"Judy",
	"Alan",
	"Theresa",
	"Ralph",
	"Beverly",
	"Gabriel",
	"Denise",
	"Roy",
	"Marilyn",
	"Juan",
	"Amber",
	"Wayne",
	"Danielle",
	"Eugene",
	"Abigail",
	"Logan",
	"Brittany",
	"Randy",
	"Rose",
	"Louis",
	"Diana",
	"Russell",
	"Natalie",
	"Vincent",
	"Sophia",
	"Philip",
	"Alexis",
	"Bobby",
	"Lori",
	"Johnny",
	"Kayla",
	"Bradley",
	"Jane",
}

// https://www.paperrater.com/page/lists-of-adjectives
var adjectives = []string{
	"attractive",
	"bald",
	"beautiful",
	"chubby",
	"clean",
	"dazzling",
	"drab",
	"elegant",
	"fancy",
	"fit",
	"flabby",
	"glamorous",
	"gorgeous",
	"handsome",
	"long",
	"magnificent",
	"muscular",
	"plain",
	"plump",
	"quaint",
	"scruffy",
	"shapely",
	"short",
	"skinny",
	"stocky",
	"ugly",
	"unkempt",
	"unsightly",
	"aggressive",
	"agreeable",
	"ambitious",
	"brave",
	"calm",
	"delightful",
	"eager",
	"faithful",
	"gentle",
	"happy",
	"jolly",
	"kind",
	"lively",
	"nice",
	"obedient",
	"polite",
	"proud",
	"silly",
	"thankful",
	"victorious",
	"witty",
	"wonderful",
	"angry",
	"bewildered",
	"clumsy",
	"defeated",
	"embarrassed",
	"fierce",
	"grumpy",
	"helpless",
	"itchy",
	"jealous",
	"lazy",
	"mysterious",
	"nervous",
	"obnoxious",
	"panicky",
	"pitiful",
	"repulsive",
	"scary",
	"thoughtless",
	"uptight",
	"worried",
	"big",
	"colossal",
	"fat",
	"gigantic",
	"great",
	"huge",
	"immense",
	"large",
	"little",
	"mammoth",
	"massive",
	"microscopic",
	"miniature",
	"petite",
	"puny",
	"scrawny",
	"short",
	"small",
	"tall",
	"teeny",
	"tiny",
}

// https://owlcation.com/stem/list-of-ocean-animals
var ocean_animals = []string{
	"anemone",
	"barnacle",
	"crab",
	"dolphin",
	"eel",
	"fugu",
	"great white shark",
	"hapuka",
	"icefish",
	"jellyfish",
	"killer whale",
	"lobster",
	"mollusk",
	"narwhal",
	"nautilus",
	"octopus",
	"plankton",
	"quahog",
	"rockfish",
	"salmon",
	"sawfish",
	"seahorse",
	"seal",
	"stringray",
	"shrimp",
	"sponge",
	"starfish",
	"swordfish",
	"toadfish",
	"umbrella squid",
	"viperfish",
	"walrus",
	"xiphosura",
	"yellowtail damselfish",
	"zooplankton",
}
