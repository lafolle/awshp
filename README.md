`awshp` gets all the instances in AWS stack and populates /etc/hosts file accordingly. Duplicate entries are ignored. Entries are only added to `/etc/hosts` and no alteration is made to existing entries. `sudo` is required as writing to `/etc/hosts` is restricted to normal user (unless it has been customized).

## Install
`go get github.com/lafolle/awshp`

## Usage
`sudo awshp -region us-east2 -stackId 30984-309485-345-sldjf`
