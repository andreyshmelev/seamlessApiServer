package main

import (
	api "seamlessServer/seamlessApi"
)

func main() {
	/*
		in := bufio.NewReader(os.Stdin)
		out := bufio.NewWriter(os.Stdout)
		defer out.Flush()

			var str string
			fmt.Fscan(in, &str)

			fmt.Fprintln(out, "init git repo*****", str, "\n")
	*/
	api.NewServer()
}
