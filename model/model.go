package model

type Col []int
type Location [2]int

type Bike struct {
    CurrentLocation [2]int     `json:"currentLocation"`
    Trail           []Location `json:"trail"`
    Direction       string     `json:"direction"`
    PlayerId        int        `json:"playerId"`
    Distance        int
}

type Player struct {
    Id    int    `json:"id"`
    Name  string `json:"name"`
    Color string `json:"color"`
    Frags int    `json:"frags"`
    Owned int    `json:"owned"`
}

type Tick struct {
    Type     string `json:"type"`
    Casualty int    `json:"casualty"`
    Fragger  int    `json:"fragger"`
}

type Grid struct {
    Height int        `json:"height"`
    Width  int        `json:"width"`
    Tiles  []Col      `json:"tiles"`
    Bikes  []Bike     `json:"bikes"`
    Spawns []Location `json:"spawns"`
}

type JoinMessage struct {
    Name           string `json:"name"`
    MQTTClientName string `json:"mqttClientName"`
}

type PlayerMessage struct {
    Id              int      `json:"id"`
    Name            string   `json:"name"`
    SecretUserToken string   `json:"secretUserToken"`
    Position        Location `json:"position"`
}

type SteerCommand struct {
    Course      string `json:"course"`
    PlayerToken string `json:"playerToken"`
}

type Cardinal uint8

const (
    N = Cardinal(iota)
    E 
    S 
    W 
)

var CardinalStringMap = map[Cardinal]string{
    N: "N",
    E: "E",
    S: "S",
    W: "W",
}
