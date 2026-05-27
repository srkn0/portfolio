package server

import (
	"fmt"
	"os"
	"runtime"
)

const (
	ansiPrimary = "\033[38;5;33m"  // blue, matches site primary
	ansiAccent  = "\033[38;5;213m" // pink accent
	ansiMuted   = "\033[38;5;245m" // grey
	ansiBold    = "\033[1m"
	ansiReset   = "\033[0m"
)

const bannerArt = `   
   ███████ ██   ██
   ██      ██  ██
   ███████ █████
        ██ ██  ██
   ███████ ██   ██`

func printBanner(addr string, postCount, projectCount int) {
	color := isTerminal()

	if color {
		fmt.Println()
		fmt.Println(ansiPrimary + ansiBold + bannerArt + ansiReset)
		fmt.Println()
		fmt.Printf("   %sportfolio%s  %s·%s  %s%s%s  %s·%s  %s%d posts%s  %s·%s  %s%d project%s%s\n",
			ansiBold, ansiReset,
			ansiMuted, ansiReset,
			ansiAccent, runtime.Version(), ansiReset,
			ansiMuted, ansiReset,
			ansiAccent, postCount, ansiReset,
			ansiMuted, ansiReset,
			ansiAccent, projectCount, pluralS(projectCount),
			ansiReset,
		)
		fmt.Printf("   %slistening on%s %shttp://localhost%s%s\n\n",
			ansiMuted, ansiReset,
			ansiBold, addr, ansiReset,
		)
		return
	}

	fmt.Println()
	fmt.Println(bannerArt)
	fmt.Println()
	fmt.Printf("   portfolio · %s · %d posts · %d project%s\n", runtime.Version(), postCount, projectCount, pluralS(projectCount))
	fmt.Printf("   listening on http://localhost%s\n\n", addr)
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
