# dorc Helm Chart

A Helm chart for deploying DigitalOcean Registry Cleaner (dorc) as a Kubernetes CronJob.

## Prerequisites

- Kubernetes 1.21+
- Helm 3.0+
- DigitalOcean API token with registry access

## Installation

### Add Helm Repository

```bash
helm repo add dorc https://kozaktomas.github.io/digitalocean-registry-cleaner
helm repo update
```

### Basic Installation

First, create a secret with your DigitalOcean token:

```bash
kubectl create secret generic do-token --from-literal=DO_TOKEN=dop_v1_xxxxx
```

Then install the chart:

```bash
helm install dorc dorc/dorc \
  --set config.registry=my-registry \
  --set config.repositories[0]=backend \
  --set config.repositories[1]=frontend \
  --set doToken.existingSecret=do-token
```

### Using a Values File

Create a `my-values.yaml` file:

```yaml
config:
  registry: my-registry
  repositories:
    - backend
    - frontend
  keepTags: 5
  minAgeDays: 30
  protect:
    - latest
    - main
    - master
    - prod
    - production
  dryRun: false

doToken:
  existingSecret: do-token

schedule: "0 2 * * *"  # Run at 2 AM daily
```

Install with:

```bash
helm install dorc dorc/dorc -f my-values.yaml
```

### Terraform

```hcl
resource "helm_release" "dorc" {
  name       = "dorc"
  namespace  = "default"
  repository = "https://kozaktomas.github.io/digitalocean-registry-cleaner"
  chart      = "dorc"
  version    = "1.0.0"

  set {
    name  = "config.registry"
    value = "my-registry"
  }

  set_list {
    name  = "config.repositories"
    value = ["backend", "frontend"]
  }

  set {
    name  = "doToken.existingSecret"
    value = "do-token"
  }
}
```

## Configuration

### Required Parameters

| Parameter | Description |
|-----------|-------------|
| `config.registry` | DigitalOcean registry name |
| `config.repositories` | List of repository names to clean |
| `doToken.value` or `doToken.existingSecret` | DigitalOcean API token |

### Configuration Options

| Parameter | Description | Default |
|-----------|-------------|---------|
| `schedule` | Cron schedule for cleanup | `0 2 * * *` (daily at 2 AM) |
| `config.keepTags` | Number of release tags to keep | `5` |
| `config.minAgeDays` | Minimum age before deletion (days) | `30` |
| `config.protect` | List of protected tag names | `[latest,main,master,prod,production]` |
| `config.dryRun` | Enable dry-run mode (no deletions) | `false` |

### Image Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Image repository | `ghcr.io/kozaktomas/digitalocean-registry-cleaner` |
| `image.tag` | Image tag | Chart appVersion |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `imagePullSecrets` | Image pull secrets | `[]` |

### CronJob Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `cronjob.successfulJobsHistoryLimit` | Successful job history limit | `3` |
| `cronjob.failedJobsHistoryLimit` | Failed job history limit | `3` |
| `cronjob.concurrencyPolicy` | Concurrency policy | `Forbid` |
| `cronjob.restartPolicy` | Pod restart policy | `OnFailure` |
| `cronjob.backoffLimit` | Job backoff limit | `3` |
| `cronjob.activeDeadlineSeconds` | Job timeout in seconds | `600` |

### Security Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `podSecurityContext.runAsNonRoot` | Run as non-root user | `true` |
| `podSecurityContext.runAsUser` | User ID | `1000` |
| `podSecurityContext.runAsGroup` | Group ID | `1000` |
| `securityContext.allowPrivilegeEscalation` | Allow privilege escalation | `false` |
| `securityContext.readOnlyRootFilesystem` | Read-only root filesystem | `true` |
| `securityContext.capabilities.drop` | Dropped capabilities | `["ALL"]` |

### Resource Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `resources.limits.cpu` | CPU limit | `100m` |
| `resources.limits.memory` | Memory limit | `128Mi` |
| `resources.requests.cpu` | CPU request | `50m` |
| `resources.requests.memory` | Memory request | `64Mi` |

### Scheduling Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `nodeSelector` | Node selector | `{}` |
| `tolerations` | Tolerations | `[]` |
| `affinity` | Affinity rules | `{}` |

## Usage Examples

### Dry-Run Testing

Test the cleanup without actually deleting tags:

```bash
helm install dorc-test dorc/dorc \
  --set config.registry=my-registry \
  --set config.repositories[0]=backend \
  --set config.repositories[1]=frontend \
  --set config.dryRun=true \
  --set doToken.existingSecret=do-token
```

### Aggressive Cleanup Strategy

Keep fewer tags and delete older ones sooner:

```bash
helm install dorc dorc/dorc \
  --set config.registry=my-registry \
  --set config.repositories[0]=backend \
  --set config.repositories[1]=frontend \
  --set config.keepTags=3 \
  --set config.minAgeDays=7 \
  --set doToken.existingSecret=do-token
```

### Conservative Cleanup Strategy

Keep more tags and only delete very old ones:

```bash
helm install dorc dorc/dorc \
  --set config.registry=my-registry \
  --set config.repositories[0]=backend \
  --set config.repositories[1]=frontend \
  --set config.keepTags=20 \
  --set config.minAgeDays=90 \
  --set doToken.existingSecret=do-token
```

### Custom Protected Tags

Protect additional environment-specific tags:

```bash
helm install dorc dorc/dorc \
  --set config.registry=my-registry \
  --set config.repositories[0]=backend \
  --set config.repositories[1]=frontend \
  --set config.protect[0]=latest \
  --set config.protect[1]=main \
  --set config.protect[2]=prod \
  --set config.protect[3]=staging \
  --set doToken.existingSecret=do-token
```

### Weekly Cleanup Schedule

Run cleanup once a week on Sunday at 3 AM:

```bash
helm install dorc dorc/dorc \
  --set config.registry=my-registry \
  --set config.repositories[0]=backend \
  --set config.repositories[1]=frontend \
  --set schedule="0 3 * * 0" \
  --set doToken.existingSecret=do-token
```

## Manual Trigger

To manually trigger a cleanup job:

```bash
kubectl create job --from=cronjob/dorc dorc-manual
```

## Uninstallation

```bash
helm uninstall dorc
```

## Troubleshooting

### View CronJob Status

```bash
kubectl get cronjob dorc
```

### View Job History

```bash
kubectl get jobs -l app.kubernetes.io/instance=dorc
```

### View Logs

```bash
kubectl logs -l app.kubernetes.io/instance=dorc --tail=100
```

### Check Secret

```bash
kubectl get secret dorc -o yaml
```
