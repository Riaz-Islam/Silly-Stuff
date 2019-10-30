package tracker

import (
	b64 "encoding/base64"
	"encoding/xml"
	"errors"
	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/inflection"
	tk "github.com/kevingentile/fortnite-tracker"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"time"
)

var DataDir = "./data/"
var Extension = ".bin"
var GlobalSession *discordgo.Session

type Stat struct {
	Name     string `xml:"Name"`
	Matches  int    `xml:"Matches"`
	Kills    int    `xml:"Kills"`
	Wins     int    `xml:"Wins"`
	Platform string `xml:"Platform"`
}

type ServerStat struct {
	ServerId   string `xml:"ServerId"`
	ServerName string `xml:"ServerName"`
	ChannelId  string `xml:"ChannelId"`
	Stats      []Stat `xml:"Stats>Value"`
}

var Mutex *sync.Mutex
var GlobalStat map[string]*ServerStat
var listPlatforms = []string{"pc", "psn", "xbox"}

const channelID = "595428127446925312"

var trackerReqTime = time.Now()

func ReadFromFile(fileName string, obj interface{}) error {
	configFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	str, _ := b64.StdEncoding.DecodeString(string(configFile))
	configFile = []byte(str)
	err = xml.Unmarshal(configFile, obj)
	if err != nil {
		return err
	}
	return nil
}

func GetServerStats(server string) *ServerStat {
	ss := &ServerStat{}
	err := ReadFromFile(DataDir+server+Extension, ss)
	if err != nil {
		panic(err)
	}
	return ss
}

func SetServerStats(server string, ss *ServerStat) {
	f, err := xml.Marshal(ss)
	if err != nil {
		panic(err)
	}
	str := b64.StdEncoding.EncodeToString([]byte(f))
	f = []byte(str)
	err = ioutil.WriteFile(DataDir+server+Extension, f, 0644)
	if err != nil {
		panic(err)
	}
}

func getStat(name, platform string) (Stat, error) {
	username := name
	apiToken := "11dadfc6-ff3d-4a7e-b2bb-a0682c0d3cc1"

	profile, err := tk.GetProfile(platform, username, apiToken)
	trackerReqTime = time.Now()

	if err != nil {
		LogWarning("Error retrieving stat for :", name, platform)
		return Stat{}, err
	}
	var s Stat
	for _, b := range profile.LifeTimeStats {
		switch b.Key {
		case "Matches Played":
			s.Matches, _ = strconv.Atoi(b.Value)
		case "Wins":
			s.Wins, _ = strconv.Atoi(b.Value)
		case "Kills":
			s.Kills, _ = strconv.Atoi(b.Value)
		}
	}
	s.Name = name
	s.Platform = platform
	return s, nil
}

func Check(discord *discordgo.Session) string {
	GlobalSession = discord
	statDumpTime := time.Now()
	statPrintTime := time.Now()
	for {
		curTime := time.Now()
		if time.Now().Sub(statDumpTime) > 30*time.Second {
			go statDump(discord)
			LogVerbosef("Stat dump at : %v", time.Now())
			statDumpTime = time.Now()
			//statInit(discord)
		}

		nametostat := map[string]Stat{}
		Mutex.Lock()
		//gocommon.LogVerbosef("lock acquired by check1")
		for s := range GlobalStat {
			for _, stat := range GlobalStat[s].Stats {
				nametostat[stat.Name] = stat
			}
		}
		if time.Now().Sub(statPrintTime) > 1*time.Minute {
			LogVerbose("Globalstat length: ", len(GlobalStat))
			for s := range GlobalStat {
				LogVerbose("Stat for server: ", s, " total: ", len(GlobalStat[s].Stats))
				if GlobalStat[s] != nil || len(GlobalStat[s].Stats) > 0 {
					LogVerbose(*GlobalStat[s])
				}
			}
			statPrintTime = time.Now()
		}
		//gocommon.LogVerbosef("lock released by check1")
		Mutex.Unlock()

		for key, prev := range nametostat {
			time.Sleep(2 * time.Second)
			cur, err := getStat(prev.Name, prev.Platform)
			//LogInfof("updating stat, name=%s, prev=%+v, cur=%v", prev.Name, prev, cur)
			if err != nil {
				LogWarning("Error at check() - 1 :", err.Error())
			} else {
				nametostat[key] = cur
			}
		}

		Mutex.Lock()
		//gocommon.LogVerbosef("lock acquired by check2")
		for s := range GlobalStat {
			for i, stat := range GlobalStat[s].Stats {
				//gocommon.LogVerbosef("check diff: prev=%+v, cur=%+v", GlobalStat[s].Stats[i], nametostat[stat.Name])
				message := checkDiff(nametostat[stat.Name], GlobalStat[s].Stats[i])
				//gocommon.LogVerbosef("checkdiff message=%s", message)
				if GlobalStat[s].ChannelId == "" || message == "" {
					continue
				}
				_, _ = discord.ChannelMessageSend(GlobalStat[s].ChannelId, message)
				GlobalStat[s].Stats[i] = nametostat[stat.Name]
			}
		}
		//gocommon.LogVerbosef("lock released by check2")
		Mutex.Unlock()

		LogVerbose("One round of check in: ", time.Now().Sub(curTime))
		if time.Now().Sub(curTime) < 5*time.Second {
			time.Sleep(5*time.Second - time.Now().Sub(curTime))
		}
	}
}

func checkDiff(cur, prev Stat) string {
	m := cur.Matches - prev.Matches
	k := cur.Kills - prev.Kills
	w := cur.Wins - prev.Wins
	if m > 0 {
		rep := "**" + cur.Name + "** played " + verbose(m, "match", false) + " with " + verbose(k, "kill", false) + "\n"
		if w > 0 {
			rep += Highlight("AND WON "+verbose(w, "MATCH", true)+" !!!! 😱 👏 👏 👏", "http")
		}
		return rep
	}
	return ""
}

func Register(name, server string) error {
	pos := find(server, name)
	LogInfo("registering, found at pos %d", pos)
	if pos != -1 {
		return errors.New(name + " already registered 🤷‍")
	}
	//gocommon.LogVerbosef("registering, name=%s, server=%s", name, server)
	var err error
	s := Stat{}
	for _, p := range listPlatforms {
		s, err = getStat(name, p)
		if err == nil {
			break
		}
	}
	if err != nil || len(s.Platform) == 0 {
		return errors.New("Player not found")
	}
	LogInfo("name=%s, found for platform=%s", name, s.Platform)
	LogInfo("stat: %+v\n", s)

	//gocommon.LogVerbosef("register locked")
	Mutex.Lock()
	//gocommon.LogVerbosef("lock acquired by register")
	if GlobalStat[server].Stats == nil {
		GlobalStat[server].Stats = []Stat{}
	}
	//gocommon.LogVerbosef("appending... %+v", GlobalStat[server].Stats)
	GlobalStat[server].Stats = append(GlobalStat[server].Stats, s)
	LogInfo("name=%s, stat=%+v", name, GlobalStat[server].Stats)
	//gocommon.LogVerbosef("lock released by register")
	Mutex.Unlock()
	//gocommon.LogVerbosef("register unlocked")

	return nil
}

func Unregister(name, server string) error {
	pos := find(server, name)
	if pos == -1 {
		return errors.New("This player is not registered")
	}
	erase(server, pos)
	return nil
}

func Show(server string) (string, error) {
	reply := []string{}
	Mutex.Lock()
	//gocommon.LogVerbosef("lock acquired by show")
	for _, stat := range GlobalStat[server].Stats {
		reply = append(reply, stat.Name)
	}
	//gocommon.LogVerbosef("lock released by show")
	Mutex.Unlock()
	return "Registered: " + strings.Join(reply, ","), nil
}

func statInit(discord *discordgo.Session) {
	Mutex.Lock()
	//gocommon.LogVerbosef("lock acquired by stat_init")
	for _, server := range discord.State.Guilds {
		GlobalStat[server.ID] = GetServerStats(server.ID)
	}
	//gocommon.LogVerbosef("lock released by stat_init")
	Mutex.Unlock()
}

func statDump(discord *discordgo.Session) {
	Mutex.Lock()
	//gocommon.LogVerbosef("lock acquired by stat_dump")
	for _, server := range discord.State.Guilds {
		SetServerStats(server.ID, GlobalStat[server.ID])
	}
	//gocommon.LogVerbosef("lock released by stat_dump")
	Mutex.Unlock()
}

func SetChannel(channel, server string) {
	Mutex.Lock()
	//gocommon.LogVerbosef("lock acquired by set_channel")
	GlobalStat[server].ChannelId = channel
	//gocommon.LogVerbosef("lock released by set_channel")
	Mutex.Unlock()
}

func UnSetChannel(server string) {
	Mutex.Lock()
	//gocommon.LogVerbosef("lock acquired by unset_channel")
	GlobalStat[server].ChannelId = ""
	//gocommon.LogVerbosef("lock released by unset_channel")
	Mutex.Unlock()
}

func verbose(n int, s string, flag bool) string {
	if n == 0 {
		return "no " + s
	} else if n == 1 {
		if strings.ToLower(s) == "match" {
			return "a " + s
		} else {
			return "1 " + s
		}
	} else {
		return strconv.Itoa(n) + " " + inflection.Plural(s)
	}
}

func find(serverId, playerName string) int {
	Mutex.Lock()
	//gocommon.LogVerbosef("lock acquired by find")
	defer func() {
		//gocommon.LogVerbosef("lock released by find")
		Mutex.Unlock()
	}()
	for i, stat := range GlobalStat[serverId].Stats {
		if stat.Name == playerName {
			return i
		}
	}
	return -1
}

func erase(serverId string, pos int) {
	Mutex.Lock()
	//gocommon.LogVerbosef("lock acquired by erase")
	GlobalStat[serverId].Stats = append(GlobalStat[serverId].Stats[:pos], GlobalStat[serverId].Stats[pos+1:]...)
	//gocommon.LogVerbosef("lock released by erase")
	Mutex.Unlock()
}

func BroadCast(discord *discordgo.Session, message string) {
	Mutex.Lock()
	for _, server := range GlobalStat {
		if server.ChannelId != "" {
			discord.ChannelMessageSend(server.ChannelId, message)
		}
	}
	Mutex.Unlock()
}
