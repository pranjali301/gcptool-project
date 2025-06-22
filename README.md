# GCP Cloud Function Project

This repository contains:

- A Google Cloud Function (in the `function/` directory)
- A custom CLI tool (`gcptool` in `cmd/gcptool/`) for deploying and managing the function
- GitHub Actions workflow for CI/CD

---

## Prerequisites

- [Go 1.21+](https://golang.org/dl/)
- [gcloud CLI](https://cloud.google.com/sdk/docs/install) (for manual deployment)
- Google Cloud project with Cloud Functions enabled
- Service account with permissions to deploy Cloud Functions

---

## Project Structure# GCP Cloud Function Project

This repository contains:

- A Google Cloud Function (in the `function/` directory)
- A custom CLI tool (`gcptool` in `cmd/gcptool/`) for deploying and managing the function
- GitHub Actions workflow for CI/CD

---

## Prerequisites

- [Go 1.21+](https://golang.org/dl/)
- [gcloud CLI](https://cloud.google.com/sdk/docs/install) (for manual deployment)
- Google Cloud project with Cloud Functions enabled
- Service account with permissions to deploy Cloud Functions

---

## Project Structure

```
.
├── cmd/gcptool/         # CLI tool source code
├── function/            # Cloud Function source code
├── .github/workflows/   # GitHub Actions workflows
├── go.mod
└── README.md
```

---

## Local Development

### 1. Install dependencies

```sh
go mod tidy
cd cmd/gcptool
go mod tidy
```

### 2. Run tests

```sh
cd function
go test -v
```

### 3. Build the CLI tool

```sh
cd cmd/gcptool
go build -o ../../gcptool .
```

### 4. Deploy the function manually

```sh
./gcptool deploy my-cloud-function
```

Or, using `gcloud` directly:

```sh
gcloud functions deploy my-cloud-function \
  --gen2 \
  --runtime=go121 \
  --region=us-central1 \
  --source=./function \
  --entry-point=HandleRequest \
  --trigger=https \
  --allow-unauthenticated \
  --project=<YOUR_PROJECT_ID>
```

---

## GitHub Actions CI/CD

On every push or pull request to `main`:

- Tests are run (`function/`)
- The CLI tool is built (`cmd/gcptool/`)
- On push to `main`, the function is deployed using the CLI tool

**Secrets required:**

- `GCP_SA_KEY`: Service account key JSON
- `GCP_PROJECT_ID`: Your GCP project ID

---

## Testing the Deployed Function

After deployment, the workflow will:

- Fetch the function URL
- Send a test HTTP request using `curl`

---

## Environment Variables

- `GCP_PROJECT_ID`: Used by the CLI tool and workflows

---