# Security Philosophy

## Core Position
- Yanzi is local-first.
- Yanzi does not collect user data by default.
- Governance artifacts remain user-controlled.

## Operational Security Posture
- Organizations define and enforce their own security posture.
- Yanzi should remain deployment-flexible across environments.
- External secret providers are recommended when secrets are required.

## Boundary Clarification
### Governance
Defines policy, certification expectations, and review discipline.

### Secrets Management
Handled by operator/organization infrastructure; not embedded as hard assumptions in Yanzi.

### Runtime Execution
Yanzi CLI behavior should remain explicit, inspectable, and deterministic.

### Operational Ownership
Users and organizations own environment hardening, access control, and release approval decisions.

## Trust Implications
- Security trust is improved by deterministic operations and preserved provenance.
- Governance documentation complements, but does not replace, environment security controls.
