package sgf

import (
	"os"
	"strconv"
)

func SgfSize(name string) (string, error) {
	f, err := os.ReadFile(name)
	if err != nil {
		return "0,0,19,19", err
	}
	s := string(f)
	minfirst := byte('s')
	maxfirst := byte('a')
	minsecond := byte('s')
	maxsecond := byte('a')
	setblack := false
	setwhite := false
	sz := 19
	for i := 0; i < len(s)-6; i++ {
		if s[i:i+3] == "SZ[" {
			sz, err = strconv.Atoi(s[i+3 : i+5])
			if err != nil {
				sz = 19
			}
		}
		if (!setblack || !setwhite) && s[i] == 'A' && (s[i+1] == 'B' || s[i+1] == 'W') && s[i+2] == '[' {
			if s[i+1] == 'B' {
				setblack = true
			} else {
				setwhite = true
			}
			j := i + 2
			for s[j] == '[' && s[j+3] == ']' {
				if s[j+1] >= 'a' && s[j+1] <= 's' {
					minfirst = min(minfirst, s[j+1])
					maxfirst = max(maxfirst, s[j+1])
				}
				if s[j+2] >= 'a' && s[j+2] <= 's' {
					minsecond = min(minsecond, s[j+2])
					maxsecond = max(maxsecond, s[j+2])
				}
				j += 4
			}
			i = j - 1
		}
		if s[i] == ';' && (s[i+1] == 'B' || s[i+1] == 'W') {
			if s[i+2] == '[' && s[i+5] == ']' {
				if s[i+3] >= 'a' && s[i+3] <= 's' {
					minfirst = min(minfirst, s[i+3])
					maxfirst = max(maxfirst, s[i+3])
				}
				if s[i+4] >= 'a' && s[i+4] <= 's' {
					minsecond = min(minsecond, s[i+4])
					maxsecond = max(maxsecond, s[i+4])
				}
			}
		}
	}
	left, top, right, bottom := 0, 0, sz, sz
	left = max(left, int(minfirst-byte('a'))-3)
	top = max(top, int(minsecond-byte('a'))-3)
	right = min(right, int(maxfirst-byte('a'))+3)
	bottom = min(bottom, int(maxsecond-byte('a'))+3)
	return strconv.Itoa(left) + "," + strconv.Itoa(sz+1-bottom) + "," + strconv.Itoa(right) + "," + strconv.Itoa(sz-top), nil
}
