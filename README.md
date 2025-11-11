## Nanabush vLLM Platform

This directory houses the artefacts for deploying, operating, and evolving the in-cluster vLLM environment that powers the Glooscap translation operator.

### Objectives

- Provide GPU-enabled inference endpoints inside OpenShift with zero external dependencies.
- Support two interaction modes: on-demand Tekton Jobs and always-on inference services.
- Enforce strict network isolation, auditing, and sanitisation workflows to prevent data leakage.
- Maintain retraining pipelines that continuously improve the translation model using sanitized feedback.

### Proposed Structure

- `tekton/` – Pipelines, tasks, and triggers for job-oriented inference and retraining.
- `helm/vllm/` – Helm chart for persistent vLLM deployment (replicas, autoscaling, mTLS).
- `kustomize/` – Overlays for dev/test/prod namespaces.
- `policies/` – NetworkPolicies, SCCs, Compliance Operator profiles, sandbox runtime configs.
- `scripts/` – Helper scripts (`buildme.sh`, `cycleme.sh`, `deployme.sh`) aligned with org norms.

### Next Steps

1. Define GPU machine set and node selector/taints for vLLM workloads.
2. Draft Tekton pipeline (`translation-pipeline.yaml`) referencing secure volumes for model weights.
3. Scaffold Helm chart with Service Mesh annotations and OTEL sidecar.
4. Document retraining strategy and approval gates in `docs/`.
5. Integrate cluster monitoring (Prometheus, Grafana) and alerting for GPU utilisation.

Refer back to `tools/glooscap/docs/vllm-integration.md` for operator integration details.

