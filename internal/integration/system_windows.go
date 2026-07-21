//go:build windows

package integration

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/Nico-Mayer/themectl/internal/theme"
	winReg "golang.org/x/sys/windows/registry"
)

const (
	hwndBroadcast   = 0xFFFF
	wmSettingChange = 0x001A
	wmThemeChanged  = 0x031A
)

func checkSystemAppearance() error {
	return nil
}

func setSystemAppearance(appearance theme.Appearance) error {
	var value uint32
	switch appearance {
	case theme.Dark:
		value = 0
	case theme.Light:
		value = 1
	default:
		return fmt.Errorf("unsupported appearance %q: expected %q or %q", appearance, theme.Dark, theme.Light)
	}

	key, err := winReg.OpenKey(
		winReg.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`,
		winReg.SET_VALUE,
	)
	if err != nil {
		return err
	}
	defer key.Close()

	if err := key.SetDWordValue("AppsUseLightTheme", value); err != nil {
		return err
	}
	if err := key.SetDWordValue("SystemUsesLightTheme", value); err != nil {
		return err
	}

	user32 := syscall.NewLazyDLL("user32.dll")
	sendMessageTimeout := user32.NewProc("SendMessageTimeoutW")

	param, _ := syscall.UTF16PtrFromString("ImmersiveColorSet")

	sendMessageTimeout.Call(
		hwndBroadcast,
		wmSettingChange,
		0,
		uintptr(unsafe.Pointer(param)),
		0,
		100,
		0,
	)
	sendMessageTimeout.Call(
		hwndBroadcast,
		wmThemeChanged,
		0,
		0,
		0,
		100,
		0,
	)

	return nil
}
