// Package apperrors は Unit 横串の sentinel error を定義する。
// 各 Unit はここに定義された error を import して errors.Is で比較する。
package apperrors

import "errors"

// ErrUnauthorized は認証されていないリクエストに対して返される sentinel error。
// unit-interfaces.md §2.1 で公開されている契約。
var ErrUnauthorized = errors.New("unauthorized")
