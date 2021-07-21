package server

import (
	"fmt"
	"time"

	rtctokenbuilder "github.com/AgoraIO/Tools/DynamicKey/AgoraDynamicKey/go/src/RtcTokenBuilder"
)

var (
	AgoraAppID          string
	AgoraAppCertificate string
)

var agoraNextUID uint32 = 1

const agoraExpireTimeInSeconds = uint32(24 * 60 * 60)

func getAgoraToken(channelName string) (uint32, string, error) {
	if len(AgoraAppCertificate) == 0 {
		return 0, "", fmt.Errorf("agora credentials not configured")
	}
	uid := agoraNextUID
	agoraNextUID++

	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp := currentTimestamp + agoraExpireTimeInSeconds

	result, err := rtctokenbuilder.BuildTokenWithUID(AgoraAppID, AgoraAppCertificate, channelName, uid, rtctokenbuilder.RoleAttendee, expireTimestamp)
	if err != nil {
		return 0, "", err
	}
	return uid, result, nil
}
