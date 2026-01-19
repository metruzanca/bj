# Background Jobs

Simple Golang CLI util that reliably sends a job to the background. Sometimes appending `&` to the end of a command is not enough to send it to the background. (e.g. `mise install &` will not go to background for some reason).

`bj` usage will be simple: `bj mise install` will run the $cmd (in this case mise install) in a background process (maybe take the current shell and spawn a new process of the shell in the current pwd and running the command).

`bj` will do some minimal tracking, and pipe all logs to a file. Accessing these logs is as simple as `bj --list` or `bg -l` (important to note that bj will only consider flags immediately after the `bj` word as its own, all others will be ignored as they're considered as command flags). This list will show a list of all files, start and end date, execution time. These logs should live somewhere under the home directory.

`bj --logs` without args will show logs for the latest command logs, with an argument it will show that file. The file will be opened with `less`.

`bj` will have a small config file in ~/.config/bj/bj.toml this should have flat config for log location (default ~/.config/bj/logs/), logs viewer command (default less).

