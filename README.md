# dev

## Syncing the Drivers Directory:

Use `rsync`:

```bash
rsync -avz mongo-go-driver Administrator@ec2-54-81-161-56.compute-1.amazonaws.com:/cygdrive/c/data/
```

rhel:
```
rsync -avc mongo-go-driver ec2-user@ec2-3-101-57-57.us-west-1.compute.amazonaws.com:/home/ec2-user/
```

## Resources 

- [Godoc](https://go.googlesource.com/tools/+/refs/heads/master/godoc/#godoc)
- [How to modify and push to someone else's PR on github](https://gist.github.com/wtbarnes/56b942641d314522094d312bbaf33a81)
- [List of MongoDB Server error codes](https://github.com/mongodb/mongo/blob/master/src/mongo/base/error_codes.yml)

