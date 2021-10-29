# Serverless Workflow 
Why serverless workflows ? 

From [Google Cloud Blog](https://cloud.google.com/blog/products/application-development/get-to-know-google-cloud-workflows)

*If you want to process events or chain APIs in a serverless way—or have workloads that are bursty or latency-sensitive—we recommend Workflows.* 

*Workflows scales to zero when you’re not using it, incurring no costs when it’s idle. Pricing is based on the number of steps in the workflow, so you only pay if your workflow runs. And because Workflows doesn’t charge based on execution time, if a workflow pauses for a few hours in between tasks, you don’t pay for this either.*

*Workflows scale up automatically with very low startup time and no “cold start” effect. Also, it transitions quickly between steps, supporting latency-sensitive applications.*

## Proof of concept

Develop a serverless business process. 
Image processing pipeline: Manipulating images is a common required process when working with images, for websites or for building ML image models. 

### Google Cloud Services in place: 

* **Cloud Storage**
Since we want to store imgages there is no a better place to store them than Cloud Storage buckets. Moreover, we can handle events and trigger some computation from those events. 

* **CloudRun**
It is possible to reach the same goals using Cloud Functions, but one of the objectives is to demostrate some Cloud Run capabilities. 
Golang was choosed for developing the services, because it is a good candidate for Cloud Run: less coldtart, less container size (around 35Mb each container), and less reource consumption in general, since no VM is needed than other languages like Python, NodeJS or Java. That should be translated to less pricing. Performance and concurrency support are other strong points of Golang, but they are out of scope of this POC.

* **Workflows**
The ability to orchestrate the process, where we only need to define which steps will be executed to respond an event.
Workflows requires no infrastructure management and scales seamlessly with demand, including scaling down to zero. Workflows's rapid scaling and low execution delay make it a great fit for latency-sensitive implementations. A workflow is made up of a series of steps described using the Workflows syntax, which can be written in either the *YAML* or *JSON* format. Workflows come with a complete feature list: Subworkflows, Conditional Steps, Passing variable values between workflow steps, Built-in authentication for Google Cloud products, and so on. Check the full list --> [Google Cloud Workflows Features](https://cloud.google.com/workflows#all-features)

* **EventArc** 
Since Serverless services should be event oriented, they consume resources when there is an avent to process, but they can take a nap if there is no work to do, to reduce cost.
Having an Event management is relevant, EventArc lets you asynchronously deliver events from Google services, SaaS, and your own apps using loosely coupled services that react to state changes, other good point for EvenntArc is that it uses [*CloudEvents*](https://cloudevents.io/) for the event envelop, something that facilitates the integration between services. Moreover, it pushes events to the subscribers using an HTTP POST request, what reduces the need of using message brokers client libraries, since everything works as HTTP request, event processors are HTTP services.

  Eventarc Workflows destination is currently a feature in **private preview**. It will allow to connect Eventarc events to Workflows directly, for example, we can create an Eventarc Cloud Storage trigger to listen for new object creations events in a bucket and pass them onto a workflow.

### Image Processing Workflow 

The image processing workflow:

![Image Processing Workflow](/doc/img_processing.png "Workflow")

#### Services

- [Resize](resize/Readme.md): Resizes an image and save the image result to the target bucket. 

- [Editor](editor/Readme.md): Manipulate an Image and save the image result to the target bucket. 
  
  Manipulations, it is possible to combine them:
   - Grayscale
   - AdjustBrightness
   - AdjustSaturation
   - AdjustContrast
   - Blur
   - AdjustGamma  

- [Label](label/Readme.md): Classify the image using Cloud Vision API, and copy the image to the target bucket. 


![Image Processing Workflow](/doc/img_processing_gcp.png "GCP Workflow")


#### Deployment 

* Set project ID

```
gcloud config set project [YOUR-PROJECT-ID]
PROJECT_ID=$(gcloud config get-value project)
``` 

* Enable APIs

```
gcloud services enable \
  cloudbuild.googleapis.com \
  eventarc.googleapis.com \
  vision.googleapis.com \
  workflows.googleapis.com \
  workflowexecutions.googleapis.com
```

* Configure Cloud Run and EventArc

```
REGION=[REGION]

gcloud config set run/region $REGION
gcloud config set run/platform managed
gcloud config set eventarc/location $REGION
```

* Service Accounts 

```
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')

gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member serviceAccount:$PROJECT_NUMBER-compute@developer.gserviceaccount.com \
    --role roles/eventarc.eventReceiver

SERVICE_ACCOUNT="$(gsutil kms serviceaccount -p ${PROJECT_NUMBER})"

gcloud projects add-iam-policy-binding ${PROJECT_NUMBER} \
    --member serviceAccount:${SERVICE_ACCOUNT} \
    --role roles/pubsub.publisher

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member serviceAccount:service-$PROJECT_NUMBER@gcp-sa-pubsub.iam.gserviceaccount.com \
  --role roles/iam.serviceAccountTokenCreator        
```

* Create Buckets 

Input bucket 

```
BUCKET1=$PROJECT_ID-images-input 
gsutil mb -l $REGION gs://$BUCKET1
```

Ouput bucket

```
BUCKET2=$PROJECT_ID-images-output
gsutil mb -l $REGION gs://$BUCKET2
```
 
##### Deploy Services

* [Resize](resize/Readme.md)
* [Editor](editor/Readme.md)
* [Label](label/Readme.md)

##### Deploy the workflow:

```
WORKFLOW_NAME=image-processing
gcloud workflows deploy $WORKFLOW_NAME --source=workflow.yaml
```

##### EventArc 

Note: Eventarc *Workflows destination* is currently a feature in private preview. 

```
TRIGGER_NAME=trigger-$WORKFLOW_NAME
gcloud eventarc triggers create $TRIGGER_NAME \
  --location=$REGION \
  --destination-workflow=$WORKFLOW_NAME \
  --destination-workflow-location=$REGION \
  --event-filters="type=google.cloud.storage.object.v1.finalized" \
  --event-filters="bucket=$BUCKET1" \
  --service-account=$PROJECT_NUMBER-compute@developer.gserviceaccount.com
```

List triggers:

```
gcloud eventarc triggers list --location=us-central1
```

##### Workflow 

The first step is to process the EventArc event and extract the values, the bucket and the created object. 

```yaml
main:
  params: [event]
  steps:
  - log_event:
      call: sys.log
      args:
          text: ${event}
          severity: INFO
  - extract_bucket_and_file:
      assign:
      - bucket: ${event.data.bucket} #variable from event 
      - file: ${event.data.name} #variable from event
```

The Resize step:

```yaml
  - resize:
      call: http.post
      args:
        url: [service_url] #Replace
        auth:
          type: OIDC
        body:
            bucket: ${bucket} # ref
            object: ${file} # ref
            width: 480 # size
            output: [output_bucket_here] # output bucket - Replace
            outputpath: resized/480 # output path 
      result: resizeResponse #result 
``` 
Edit the resized image:

```yaml
  - edit:
      call: http.post
      args:
        url: [service_url] #Replace
        auth:
          type: OIDC
        body:
            bucket: ${resizeResponse.body.bucket}  # from the resized response
            object: ${resizeResponse.body.object}  # from the resized response
            chain: 
              AdjustBrightness: 40
              AdjustContrast: 60
            output: [output_bucket_here] # output bucket - Replace
            outputpath: edited # output path 
      result: editResponse   #result 
``` 

Classify the image: 

```yaml
  - label:
      call: http.post
      args:
        url: [service_url] #Replace
        auth:
          type: OIDC
        body:
            bucket: ${resizeResponse.body.bucket} # from the resized response
            object: ${resizeResponse.body.object} # from the resized response          
            output: [output_bucket_here] # output bucket - Replace             
      result: labelResponse 
``` 

End step: 

```yaml
  - final:
      return:
        file: ${event.data.name}
        resize: ${resizeResponse.body.object}
        edited: ${editResponse.body.object} 
        labeled: ${labelResponse.body.object} 
```

## Next

- [ ] Conditional steps
- [ ] Subworkflows
- [ ] Callbacks
- [ ] For loops
 
