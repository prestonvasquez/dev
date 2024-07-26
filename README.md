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
