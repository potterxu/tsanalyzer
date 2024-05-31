package ts_test

import (
	"bufio"
	"encoding/hex"
	"os"
	"testing"

	"github.com/Comcast/gots/v2/packet"
	"github.com/potterxu/tsanalyzer/internal/mpts/ts"
)

func TestDefaultAccumulator(t *testing.T) {
	a := ts.NewAccumulator(nil)
	t.Log("potter")

	f, _ := os.Open("data/data.ts")
	t.Log("potter")
	defer f.Close()
	reader := bufio.NewReader(f)
	buffer := make([]byte, 188)
	readyCnt := 0
	for {
		_, err := reader.Read(buffer)
		if err != nil {
			break
		}
		pkt := packet.Packet(buffer)
		result, ready, err := a.Add(pkt)
		if err != nil {
			t.Error(err)
			return
		}
		if ready {
			readyCnt++
			expected := "000001e000008780052109e5ca9b00000001093000000001060104002d80018000000001060409b500314454473141f88000000001419a2dd012e4effffffff8fb8575faffebf777c4bcf6f27f8d21e6a5acaaf59743af4927760d9bda6abd45da21a9db86db8a5797c37acf95eb239b5ef93fe9d6b8e4d6775aed61227fcc1c0f4045623c4e96fe0801002cea7fcd84ee8693af5a5e09be33e1ef8ecef1bbdf57aeebaba8668d338c1a320cbc15411ee968a97bd7c9ef08bff27fb2992a55e2bbf8bf460a5a6b6baa8baa932fdbc34c6fe84da8d48ec995fa8927fc573c5f95faf6149e6924ccd7cc2ea3fe60a1e735498ad2ad5c276752d7d9c23c0a008318e8aa2a571c5763652d43b6a1df5410ff4fd0746c7107e7bed253c17e22cf8fda5c83ebd40760907e0bef5a8c55ddeba41f5fcc2014456e0efa5eda9341efb93993ff50570fb0510c0a8738a38f874f0561657dbc913267bf5e3c958fcfd577278467c09b60c3ef871259ba36a4ff2be89845e04c0288280a05a1d3935fd3a99df533f54077dbf10a24ff820c2c1f2d5855ee8e97cfe9e4ff05c39901984e70a6dd4716f6a978f6515bb3812890eeb4c9a4802a5831df93abad50ee6a77142b8fdbac227e08e249ff2bfc9ad3fbf26117fd5559b19175ac94b0a537717cb4b4ff5fbae5f1cbdfe5dfdc25afcbffcb17c4fd099fc7e9e74bf4ff1f74afbd9bff0a5fded4c4718f28ca77fbfc766531c41dee51ae6ab8797e2bdb7777af367c6ef7d6eb0e3f12fef5fefb1ff23f22d7b709dc23aeffbdfd287f7adebc166c61a9135d59e8299bd6d0cbdc78457fad3cdcbf7ae5bb87f7ae59ee2f5857b8cb8195f88ebfd7f93ffc4237c475febfe087cdf11f37c475febfd7faffcdf11df9be23aff5febfd7fe6828e34093e1e977d7de23aff5febff345f4c8ecae003aace97185bfe15dedefcbe4b0d0d703ef"
			str := hex.EncodeToString(result.Data)
			if expected != str {
				t.Error("result not matched")
			}
		}
	}
	if readyCnt != 1 {
		t.Error("no complete pes found")
	}
}
