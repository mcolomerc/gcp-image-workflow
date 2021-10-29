# RESIZE SERVICE

Service to resize images. 

It takes a source image from Cloud Storage, resizes it to a given width using Lanczos3 and maintaining proportion.

Request Body: 

 * bucket: Source bucket
 * object: Source object 
 * width: Width size 
 * output: Output bucket 
 * outputpath: Output path 

# BUILD AND DEPLOY

* Google Cloud Project configuration

```
export PROJECT_NUMBER="$(gcloud projects describe $(gcloud config get-value project) --format='value(projectNumber)')" 
```

```
export PROJECT=[YOUR-PROJECT-ID]
```

``` 
export SERVICE="resize-service" 
```

``` 
export REGION="[YOUR_REGION]"
```

``` 
export IMAGE_URL="gcr.io/$(gcloud config get-value project)/resize-service" 
``` 

* IAM

``` 
gcloud projects add-iam-policy-binding $(gcloud config get-value project) \
    --member="serviceAccount:service-${PROJECT_NUMBER}@gcp-sa-pubsub.iam.gserviceaccount.com"\
    --role='roles/iam.serviceAccountTokenCreator' 
```

* Enabling Cloud Build

```
gcloud services enable cloudbuild.googleapis.com
```

* Enabling Cloud Run

```
gcloud services enable run.googleapis.com
```     

* Cloud Run configuration

```
gcloud config set run/platform managed
```

```
gcloud config set run/region ${REGION}
```

* BUILD
  
```
gcloud builds submit --tag ${IMAGE_URL} 
```  

``` 
gcloud beta run deploy ${SERVICE} --image ${IMAGE_URL} --platform managed --no-allow-unauthenticated
```

## Next

- [ ] Support more configurations
 
