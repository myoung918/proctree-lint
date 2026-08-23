# proctree

A validating parser and pretty printer for process tree snapshots.

I keep ending up with flat lists of pid/ppid/command triples pulled from
`/proc`, from a container runtime, or copied out of a core dump - and
wanting to look at them as a tree without piping them through `ps` on
whatever machine they came from. This is a small, dependency-free tool
for that: it reads the triples, checks that they actually describe a
tree (no duplicate pids, no dangling parents, no cycles), and prints
the result indented like `pstree`.

## Format

One process per line:

```
pid ppid command
```

- `pid` is a positive integer, unique within the file.
- `ppid` is a non-negative integer. `0` means "no parent" (a root).
  Any other value must match some other line's `pid`.
- `command` is everything after the second space, verbatim - it can
  contain spaces itself.
- Blank lines and lines starting with `#` are ignored.

Example (`example.proctree`):

```
# captured from a sandboxed build step
1 0 init
2 1 bash
3 2 make -j4
4 3 cc -c main.c
5 3 cc -c util.c
6 1 sshd
```

## Usage

From a file:

```
$ go run . example.proctree
1 init
├─ 2 bash
│  └─ 3 make -j4
│     ├─ 4 cc -c main.c
│     └─ 5 cc -c util.c
└─ 6 sshd
```

From stdin, which is the common case when piping a live snapshot
straight into the tool:

```
$ awk '{print $1, $2, $3}' /proc/*/stat 2>/dev/null | go run .
```

or explicitly with `-`:

```
$ cat example.proctree | go run . -
```

Pass `-json` to get the tree as JSON instead, e.g. for feeding into `jq`
or another tool:

```
$ go run . -json example.proctree
[
  {
    "pid": 1,
    "ppid": 0,
    "command": "init",
    "children": [
      {
        "pid": 2,
        "ppid": 1,
        "command": "bash",
        "children": [
          {
            "pid": 3,
            "ppid": 2,
            "command": "make -j4",
            "children": [
              {
                "pid": 4,
                "ppid": 3,
                "command": "cc -c main.c"
              },
              {
                "pid": 5,
                "ppid": 3,
                "command": "cc -c util.c"
              }
            ]
          }
        ]
      },
      {
        "pid": 6,
        "ppid": 1,
        "command": "sshd"
      }
    ]
  }
]
```

Validation failures point at the offending line:

```
$ printf '1 0 init\n2 1 bash\n2 1 vim\n' | go run .
proctree: line 3: duplicate pid 2 (first seen at line 2)
```

## Building

```
go build -o proctree .
```

No third-party dependencies - standard library only.
