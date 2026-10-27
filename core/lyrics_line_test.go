package core

import "testing"

func TestLineSyncedLyrics(t *testing.T) {
	for _, tc := range []struct{ name, input, want string }{
		{"square cues", "[00:01.00]你[00:01.50]好[00:02.00]", "[00:01.00]你好"},
		{"angle cues", "[00:12.340]<00:12.340>你<00:12.800>好<00:13.000>", "[00:12.340]你好"},
		{"repeated lines", "[00:01.00][01:02.30]再次唱起", "[00:01.00][01:02.30]再次唱起"},
		{"metadata and translations", "\ufeff[ti:歌名]\r\n[offset:-200]\r\n[00:01.00]你[00:01.50]好\r\n[00:01.00]Hello\r\n[00:03.00]\r\n", "[ti:歌名]\n[offset:-200]\n[00:01.00]你好\n[00:01.00]Hello\n[00:03.00]\n"},
		{"plain lyrics", "没有时间的歌词\n第二行", "没有时间的歌词\n第二行"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := LineSyncedLyrics(tc.input); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
