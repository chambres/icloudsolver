package main

import (
	"fmt"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/micmonay/keybd_event"
	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/sys/windows"
)

var (
	user32                   = syscall.NewLazyDLL("user32.dll")
	enumWindows              = user32.NewProc("EnumWindows")
	enumChildWindows         = user32.NewProc("EnumChildWindows")
	getWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	isWindowVisible          = user32.NewProc("IsWindowVisible")
	getWindowTextW           = user32.NewProc("GetWindowTextW")
	getWindowTextLengthW     = user32.NewProc("GetWindowTextLengthW")
	getClassNameW            = user32.NewProc("GetClassNameW")
)

func typeString(s string) {
	log(fmt.Sprintf("Typing %d characters: %s", len(s), s))

	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		log(fmt.Sprintf("Failed to create keyboard: %v", err))
		return
	}

	// Need small delay for keybd_event to initialize on Windows
	time.Sleep(100 * time.Millisecond)

	for _, char := range s {
		var vk int
		switch char {
		case '0':
			vk = keybd_event.VK_0
		case '1':
			vk = keybd_event.VK_1
		case '2':
			vk = keybd_event.VK_2
		case '3':
			vk = keybd_event.VK_3
		case '4':
			vk = keybd_event.VK_4
		case '5':
			vk = keybd_event.VK_5
		case '6':
			vk = keybd_event.VK_6
		case '7':
			vk = keybd_event.VK_7
		case '8':
			vk = keybd_event.VK_8
		case '9':
			vk = keybd_event.VK_9
		default:
			continue
		}

		kb.SetKeys(vk)
		err = kb.Launching()
		if err != nil {
			log(fmt.Sprintf("Failed to type %c: %v", char, err))
		}
		kb.Clear()
		time.Sleep(100 * time.Millisecond)
	}
	log("Typing complete")
}

func log(msg string) {
	fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05.000"), msg)
}

func getWindowText(hwnd uintptr) string {
	length, _, _ := getWindowTextLengthW.Call(hwnd)
	if length == 0 {
		return ""
	}
	buf := make([]uint16, length+1)
	getWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), length+1)
	return windows.UTF16ToString(buf)
}

func getClassName(hwnd uintptr) string {
	buf := make([]uint16, 256)
	getClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 256)
	return windows.UTF16ToString(buf)
}

func extractCodeFromDialog(hwnd uintptr) string {
	log(fmt.Sprintf("Reading dialog HWND=%d...", hwnd))
	var texts []string

	callback := syscall.NewCallback(func(childHwnd uintptr, lparam uintptr) uintptr {
		text := getWindowText(childHwnd)
		if text != "" {
			texts = append(texts, text)
		}
		return 1 // Continue enumeration
	})

	enumChildWindows.Call(hwnd, callback, 0)

	log(fmt.Sprintf("Found %d text controls", len(texts)))

	allText := strings.Join(texts, " ")
	if len(allText) > 100 {
		log(fmt.Sprintf("All text: %s...", allText[:100]))
	} else {
		log(fmt.Sprintf("All text: %s", allText))
	}

	// Match codes like "612 062" or "612062"
	re := regexp.MustCompile(`\b(\d{3}\s?\d{3})\b`)
	matches := re.FindStringSubmatch(allText)
	if len(matches) > 0 {
		code := strings.ReplaceAll(matches[0], " ", "")
		log(fmt.Sprintf("[CODE FOUND] %s", code))

		// Type the code
		log("Typing code...")
		time.Sleep(500 * time.Millisecond)
		typeString(code)
		log("Code typed!")
		return code
	}

	log("[NO CODE] No 6-digit code found")
	return ""
}

func findDialogForPID(pid int32) uintptr {
	var dialogHwnd uintptr

	callback := syscall.NewCallback(func(hwnd uintptr, lparam uintptr) uintptr {
		var windowPid uint32
		getWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&windowPid)))

		if int32(windowPid) == pid {
			className := getClassName(hwnd)
			visible, _, _ := isWindowVisible.Call(hwnd)
			if className == "#32770" && visible != 0 {
				dialogHwnd = hwnd
				return 0 // Stop enumeration
			}
		}
		return 1 // Continue enumeration
	})

	enumWindows.Call(callback, 0)
	return dialogHwnd
}

func hasVisibleWindow(pid int32) bool {
	found := false

	callback := syscall.NewCallback(func(hwnd uintptr, lparam uintptr) uintptr {
		var windowPid uint32
		getWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&windowPid)))

		if int32(windowPid) == pid {
			visible, _, _ := isWindowVisible.Call(hwnd)
			if visible != 0 {
				found = true
				return 0 // Stop enumeration
			}
		}
		return 1 // Continue enumeration
	})

	enumWindows.Call(callback, 0)
	return found
}

func monitorProcess(processName string, interval time.Duration) {
	wasVisible := make(map[int32]bool)

	log(fmt.Sprintf("Monitoring %s...", processName))
	log("Press Ctrl+C to stop.\n")

	for {
		procs, err := process.Processes()
		if err != nil {
			time.Sleep(interval)
			continue
		}

		for _, proc := range procs {
			name, err := proc.Name()
			if err != nil {
				continue
			}

			if name == processName {
				pid := proc.Pid
				isVisible := hasVisibleWindow(pid)

				if prev, exists := wasVisible[pid]; exists {
					if !prev && isVisible {
						log(fmt.Sprintf("[DETECTED] PID %d: Background -> App", pid))
						dialog := findDialogForPID(pid)
						if dialog != 0 {
							extractCodeFromDialog(dialog)
						} else {
							log("Dialog not found")
						}
					} else if prev && !isVisible {
						log(fmt.Sprintf("[DETECTED] PID %d: App -> Background", pid))
					}
				} else {
					status := "Background"
					if isVisible {
						status = "App"
					}
					log(fmt.Sprintf("[FOUND] PID %d: Currently %s", pid, status))
				}

				wasVisible[pid] = isVisible
			}
		}

		// Clean up dead processes
		activePids := make(map[int32]bool)
		for _, proc := range procs {
			activePids[proc.Pid] = true
		}
		for pid := range wasVisible {
			if !activePids[pid] {
				delete(wasVisible, pid)
			}
		}

		time.Sleep(interval)
	}
}

func main() {
	monitorProcess("iCloudPasswordsExtensionHelper.exe", 500*time.Millisecond)
}
