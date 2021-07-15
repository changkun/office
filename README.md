# office

Is Changkun in the Office?

https://changkun.de/s/working?

## Usage

```
$ ./office --help
office is a command that exposes Changkun's office status to the
public. The status can be fetched via: https://changkun.de/x/working?

Version: 0.1.0
GoVersion: devel go1.17-c95464f0ea Sun Jun 27 05:06:30 2021 +0000

Command line usage:

$ office [-vacation <time>]

options:

office -vacation
        Vacation mode

examples:

office
        $ curl -L https://changkun.de/s/working
        Yes!
        $ curl -L https://changkun.de/s/working
        No, he left 10s ago.

office -vacation 2021-08-11

        $ curl -L https://changkun.de/s/working
        No, he is on vacation and will return on 11 Aug.
```

## License

MIT &copy; [Changkun Ou](https://changkun.de)