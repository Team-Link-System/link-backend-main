package util

import (
	"time"
)

// ParseKst 함수: 주어진 시간을 KST로 변환
func ParseKst(t time.Time) time.Time {
	kst, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		// KST 로드 실패 시 기본 값 반환
		return time.Time{}
	}

	// KST로 변환
	if t.Location() != kst {
		t = t.In(kst)
	}

	return t
}

// ParseUtc 함수: 주어진 시간을 UTC로 변환
func ParseUtc(t time.Time) time.Time {
	utc, err := time.LoadLocation("UTC")
	if err != nil {
		// UTC 로드 실패 시 기본 값 반환
		return time.Time{}
	}

	// UTC로 변환
	if t.Location() != utc {
		t = t.In(utc)
	}

	return t
}
