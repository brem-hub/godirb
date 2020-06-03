# Godirb


## Website bruteforce using Godirb is fast, pure and cool!
Do you play CTF or pentest websites and you need to bruteforce directories?
Are you tired of waiting when dirsearch finishes?

__Try Godirb!__

`It is fast.`

`It is cool.`

`It is Go.`

# Let`s GO
## INSTALLATION
1. Clone repository
2. Use `godurb`

## QUICK START
    godirb -u <url> -- will provide you with fast search with default wordlist [6313 words]
    godirb -h       -- will show you all flags
## FLAGS
__REQUIRED__

      godirb -u <url> -- specify url
__ADDITIONAL__

    godirb -cd <path> -- specify custom wordlist
    godirb -clear     -- clear log/ folder from all logs
    godirb -d         -- specify depth of recursion, default 1: <url>/root/
    godirb -e         -- specify extensions to use, default none
    godirb -f         -- specify file to write output
    godirb -m         -- specify http method to use [GET, POST, HEAD]
    godirb -go         -- specify amount of goroutines to use * 10, default 1 -> 1 * 10 = 10 goroutines
    godirb -p        -- specify protocol to use [HTTPS, HTTP], default HTTP
    godirb -t         -- specify time delay between requests. Use if connection is poor
    godirb -v         -- show all requests
## WORDLIST CREATION
If you want to create your own wordlist you have to follow the pattern
- one keyword per line
- if you want to use extensions like test.txt\test.php you have to add %EXT%
### Example
    admin
    test
    lol
    with_ext%EXT%
    long/shot
    long%EXT%/shot
### Credits
This is an educational project for MSHP Golang course 2020.

Student: Kulikov Bogdan

