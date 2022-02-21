package api

import (
	"encoding/base64"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io"
	"math"
	"net/http"
	"net/url"
)

var (
	deviceId        = "IOS"
	clientVersion   = uint32(37)
	apiFirstContact = "https://www.auxbrain.com/ei/first_contact"
)

type EBData struct {
	SoulFood      int32   `json:"soulFood"`
	ProphecyBonus int32   `json:"prophecyBonus"`
	SoulEggs      float64 `json:"soulEggs"`
	ProphecyEggs  int32   `json:"prophecyEggs"`
}

func GetEBandSE(eiUID string) (string, string, error) {
	responseBody := new(FirstContact)
	authenticatedMsg := new(AuthenticatedMessage)
	payload := FirstContactRequestPayload{
		EiUserId:      eiUID,
		DeviceId:      deviceId,
		ClientVersion: clientVersion,
	}

	reqBin, err := proto.Marshal(proto.Message(&payload))
	if err != nil {
		return "", "", err
	}

	reqDataEncoded := base64.StdEncoding.EncodeToString(reqBin)

	resp, err := http.PostForm(apiFirstContact, url.Values{"data": {reqDataEncoded}})
	defer func() {
		resp.Body.Close()
	}()
	if err != nil {
		return "", "", err
	}

	body, err := io.ReadAll(resp.Body)

	enc := base64.StdEncoding
	buf := make([]byte, enc.DecodedLen(len(body)))
	decoded, err := enc.Decode(buf, body)

	if err = proto.Unmarshal(buf[:decoded], authenticatedMsg); err != nil {
		return "", "", err
	}

	if err = proto.Unmarshal(authenticatedMsg.Message, responseBody); err != nil {
		return "", "", err
	}

	progress := responseBody.Data.GetProgress()

	ebData := EBData{
		SoulEggs:     progress.GetSoulEggs(),
		ProphecyEggs: progress.GetProphecyEggs(),
	}

	er := progress.GetEpicResearches()
	for _, research := range er {
		if research.Id == "soul_eggs" {
			ebData.SoulFood = research.Level
		}
		if research.Id == "prophecy_bonus" {
			ebData.ProphecyBonus = research.Level
		}
	}

	return calculateEB(ebData), forPeople(ebData.SoulEggs), nil
}

func calculateEB(data EBData) string {
	sePercent := (0.1 + (float64(data.SoulFood) * .01)) * 100
	pePercent := math.Pow(float64(1)+0.05+(float64(data.ProphecyBonus)*0.01), float64(data.ProphecyEggs)) * 100
	bonus := (sePercent * pePercent) / 100

	return forPeople(bonus * data.SoulEggs)
}

func forPeople(bigAssNumber float64) string {
	units := []string{"", "k", "m", "b", "T", "q", "Q", "s", "S", "o", "N", "d"}
	k := float64(1000)
	magnitude := math.Floor(math.Log(bigAssNumber) / math.Log(k))
	return fmt.Sprintf("%.3f%s", bigAssNumber/(math.Pow(k, magnitude)), units[int(magnitude)])
}
