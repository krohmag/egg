package api

import (
	"context"
	"egg/datastore"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"

	"github.com/golang/protobuf/proto"
)

var (
	deviceId        = "IOS"
	clientVersion   = uint32(37)
	apiFirstContact = "https://www.auxbrain.com/ei/first_contact"
)

func GetBackupFromAPI(eiUID string) (*FirstContact_Payload, error) {
	responseBody := new(FirstContact)
	authenticatedMsg := new(AuthenticatedMessage)
	payload := FirstContactRequestPayload{
		EiUserId:      eiUID,
		DeviceId:      deviceId,
		ClientVersion: clientVersion,
	}

	reqBin, err := proto.Marshal(proto.Message(&payload))
	if err != nil {
		return &FirstContact_Payload{}, err
	}

	reqDataEncoded := base64.StdEncoding.EncodeToString(reqBin)

	resp, err := http.PostForm(apiFirstContact, url.Values{"data": {reqDataEncoded}})
	defer func() {
		_ = resp.Body.Close()
	}()
	if err != nil {
		return &FirstContact_Payload{}, err
	}

	body, err := io.ReadAll(resp.Body)

	enc := base64.StdEncoding
	buf := make([]byte, enc.DecodedLen(len(body)))
	decoded, err := enc.Decode(buf, body)

	if err = proto.Unmarshal(buf[:decoded], authenticatedMsg); err != nil {
		return &FirstContact_Payload{}, err
	}

	if err = proto.Unmarshal(authenticatedMsg.Message, responseBody); err != nil {
		return &FirstContact_Payload{}, err
	}

	return responseBody.Data, nil
}

func AddUserToDatabase(ctx context.Context, store datastore.Database, backup *FirstContact_Payload, discordName string) (datastore.User, error) {
	var soulFood int32
	var prophecyBonus int32
	for _, research := range backup.GetProgress().GetEpicResearches() {
		if research.Id == "soul_eggs" {
			soulFood = research.Level
		}
		if research.Id == "prophecy_bonus" {
			prophecyBonus = research.Level
		}
	}
	user := datastore.User{
		EggIncID:        backup.EiUserId,
		DiscordName:     discordName,
		GameAccountName: backup.UserName,
		SoulFood:        soulFood,
		ProphecyBonus:   prophecyBonus,
		SoulEggs:        backup.GetProgress().GetSoulEggs(),
		ProphecyEggs:    backup.GetProgress().GetProphecyEggs(),
	}

	tx, err := store.Transaction(ctx)
	if err != nil {
		return datastore.User{}, err
	}
	defer func() {
		if err == nil {
			_ = tx.Commit()
		} else {
			_ = tx.Rollback()
		}
	}()

	record, err := tx.CreateOrUpdateUser(user)
	if err != nil {
		return datastore.User{}, err
	}

	return record, nil
}

func GetEBandSE(user datastore.User) (string, string, error) {
	// responseBody := new(FirstContact)
	// authenticatedMsg := new(AuthenticatedMessage)
	// payload := FirstContactRequestPayload{
	// 	EiUserId:      eiUID,
	// 	DeviceId:      deviceId,
	// 	ClientVersion: clientVersion,
	// }
	//
	// reqBin, err := proto.Marshal(proto.Message(&payload))
	// if err != nil {
	// 	return "", "", err
	// }
	//
	// reqDataEncoded := base64.StdEncoding.EncodeToString(reqBin)
	//
	// resp, err := http.PostForm(apiFirstContact, url.Values{"data": {reqDataEncoded}})
	// defer func() {
	// 	resp.Body.Close()
	// }()
	// if err != nil {
	// 	return "", "", err
	// }
	//
	// body, err := io.ReadAll(resp.Body)
	//
	// enc := base64.StdEncoding
	// buf := make([]byte, enc.DecodedLen(len(body)))
	// decoded, err := enc.Decode(buf, body)
	//
	// if err = proto.Unmarshal(buf[:decoded], authenticatedMsg); err != nil {
	// 	return "", "", err
	// }
	//
	// if err = proto.Unmarshal(authenticatedMsg.Message, responseBody); err != nil {
	// 	return "", "", err
	// }
	//
	// progress := responseBody.Data.GetProgress()
	//
	// ebData := datastore.User{
	// 	DiscordName:     discordName,
	// 	GameAccountName: responseBody.Data.GetUserName(),
	// 	SoulEggs:        progress.GetSoulEggs(),
	// 	ProphecyEggs:    progress.GetProphecyEggs(),
	// }
	//
	// er := progress.GetEpicResearches()
	// for _, research := range er {
	// 	if research.Id == "soul_eggs" {
	// 		ebData.SoulFood = research.Level
	// 	}
	// 	if research.Id == "prophecy_bonus" {
	// 		ebData.ProphecyBonus = research.Level
	// 	}
	// }
	//
	// fmt.Println(ebData)

	return calculateEB(user), forPeople(user.SoulEggs), nil
}

func calculateEB(data datastore.User) string {
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
