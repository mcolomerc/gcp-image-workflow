# EDITOR SERVICE 

It takes a source image from Cloud Storage, manipulates it and store the result into a given target bucket.

Image manipulation:

 - Grayscale
 - AdjustBrightness
 - AdjustSaturation
 - AdjustContrast
 - Blur
 - AdjustGamma  

It is possible to combine manipulations. 

Example:
    AdjustBrightness: 40
    AdjustContrast: 60 

Request Body: 

* bucket: Source bucket

* object: Object path 

* chain: Manipulation map
 
* output: Output bucket 

* outputpath: Output path 

## BUILD AND DEPLOY

* Google Cloud Project configuration

```
export PROJECT_NUMBER="$(gcloud projects describe $(gcloud config get-value project) --format='value(projectNumber)')" 
```

```
export PROJECT=[YOUR-PROJECT-ID]
```

``` 
export SERVICE="editor-service" 
```

``` 
export REGION=[REGION]
```

``` 
export IMAGE_URL="gcr.io/$(gcloud config get-value project)/editor-service" 
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

* BUILD & DEPLOY
  
```
gcloud builds submit --tag ${IMAGE_URL} 
```  

``` 
gcloud beta run deploy ${SERVICE} --image ${IMAGE_URL} --platform managed --no-allow-unauthenticated
``` 

## Next

- [ ] Support more image manipulations