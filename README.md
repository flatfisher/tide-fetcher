# tide-loader
Acquire tide data and store it in Firestore

## Deploy

```
$ gcloud app deploy app.yaml
$ gcloud app deploy queue.yaml
```

## Scheduler

```
$ gcloud beta scheduler jobs create app-engine save-tide --schedule "30 00 * * *" --time-zone Asia/Tokyo --description "save tide to firestore" --service tide-fetcher --relative-url `/v1/tide/tasks`
```