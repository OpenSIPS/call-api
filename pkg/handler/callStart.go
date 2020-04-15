//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.
//

package handler

import (
	"fmt"
	"errors"
	"strings"
	"strconv"
)

func (h *Handler) CallStart(params []string) (string, error) {

	const headersFormat = "From: <%s>\r\n" +
		"To: <%s>\r\n" +
		"Contact: <%s>\r\n" +
		"Content-Type: application/sdp\r\n"

	const inviteBody = "v=0\r\n" +
		"o=click-to-dial 0 0 IN IP4 0.0.0.0\r\n" +
		"s=session\r\n" +
		"c=IN IP4 0.0.0.0\r\n" +
		"t=0 0\r\n" +
		"m=audio 9 RTP/AVP 0\r\n" +
		"a=rtpmap:0 PCMU/8000\r\n"

	const referHeadersFormat = "Referred-By: %s\r\n" +
		"Refer-To:: <%s>\r\n"

	if len(params) < 2 {
		return "", errors.New("caller and/or callee not specified")
	}
	caller := params[0]
	callee := params[1]

	headers := fmt.Sprintf(headersFormat, caller, callee, caller)

	var inviteParams = map[string]string{
		"method": "INVITE",
		"ruri": caller,
		"headers": headers,
		"body": inviteBody,
	}

	ret, err := h.mi.Call("t_uac_dlg", &inviteParams)
	if err != nil {
		return "", err
	}
	status, ok := ret["Status"].(string)
	if ok != true {
		return "", errors.New("invalid returned status type")
	}

	if strings.Split(status, " ")[0] != "200" {
		return "", errors.New(status)
	}

	ruri, ok := ret["RURI"].(string)
	if ok != true {
		return "", errors.New("invalid RURI returned")
	}

	message, ok := ret["Message"].(string)
	if ok != true {
		return "", errors.New("invalid Message returned")
	}

	headers = ""
	cseq := 1
	callid := ""
	for _, header := range strings.Split(message, "\r\n") {
		switch strings.Split(header, ":")[0] {
		case "CSeq":
			cseqInt, err := strconv.Atoi(strings.Split(header, " ")[1])
			if err == nil {
				cseq = cseqInt
			}
		case "Call-ID":
			callid = strings.TrimSpace(strings.Split(header, ":")[1])
			fallthrough
		case "From", "To", "Routes":
			headers += header + "\r\n"
		}
	}

	referHeaders := headers + "CSeq: " + strconv.Itoa(cseq + 1) + " REFER\r\n" +
					fmt.Sprintf(referHeadersFormat, caller, callee)
	var referParams = map[string]string{
		"method": "REFER",
		"ruri": ruri,
		"headers": referHeaders,
	}


	ret, err = h.mi.Call("t_uac_dlg", &referParams)
	if err != nil {
		return "", err
	}

	status, ok = ret["Status"].(string)
	if ok != true {
		return "", errors.New("invalid returned status type")
	}

	status = strings.Split(status, " ")[0]

	byeHeaders := headers + "CSeq: " + strconv.Itoa(cseq + 2) + " BYE\r\n"
	var byeParams = map[string]string{
		"method": "BYE",
		"ruri": ruri,
		"headers": byeHeaders,
	}
	h.mi.Call("t_uac_dlg", &byeParams)


	if status != "202" {
		return "", errors.New("could not redirect call " + callid)
	} else {
		return "call successfully started call " + callid, nil
	}
}
