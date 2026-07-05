# Security Policy

## Supported Versions

We currently provide security updates for the version `1.0.0` of this project

## Reporting a Vulnerability

We take the security of this project seriously. **Please do not report security vulnerabilities through public GitHub issues.** 

Instead, please report them by reaching out privately:

1. **Email:** Send a description of the issue to `georgeallen378@gmail.com`.
2. **GitHub Security Advisory:** You can report vulnmerablities directly via the "Security" tab.

### What to expect

* **Acknowledgment:** We will acknowledge receipt of your vulnerability report within `48 hrs`.
* **Triage:** We will investigate the issue and determine its validity and severity.
* **Resolution:** If the vulnerability is confirmed, we will work on a patch. We will keep you informed of our progress and coordinate a public disclosure timeline with you.

We ask that you maintain confidentiality until we have released a fix.

## Scope

This policy applies specifically to the code maintained in this repository. Vulnerabilities found in third-party upstream dependencies should be reported directly to those respective maintainers, though we appreciate a heads-up so we can bump our dependency versions.

## Network checks and SSRF considerations

`http_reachable` and `tcp_reachable` open outbound connections to hosts specified in `preboot.yml`. This is intentional — preboot is a local developer-environment health tool and these checks exist to confirm that required services are reachable from the developer's machine.

**What this means in practice:**

- Preboot does **not** maintain a denylist of private or loopback ranges. Connections to `127.0.0.1`, `10.x.x.x`, `192.168.x.x`, and similar addresses are allowed and expected (local dev services are the primary use case).
- Because there is no denylist, a malicious `preboot.yml` committed to a repository could direct `http_reachable`/`tcp_reachable` checks to arbitrary hosts and trigger requests to internal services.
- Redirect following is disabled — `http_reachable` does not silently follow 3xx responses to other hosts.

**Mitigations expected from users:**

1. **Treat `preboot.yml` as trusted code.** Review it in code review the same way you would a `Makefile` or CI configuration.
2. **Do not run untrusted `preboot.yml` files** from repositories you do not control.

## `port_free` loopback scope

`port_free` binds `127.0.0.1:<port>` to test availability. It only checks the loopback interface. A service listening exclusively on a non-loopback interface will not be detected — the port will appear free even when the service is running.
