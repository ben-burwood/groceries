export class AuthRedirectError extends Error {
    constructor() {
        super("auth redirect");
        this.name = "AuthRedirectError";
    }
}

let reloading = false;

export async function apiFetch(input: RequestInfo | URL, init: RequestInit = {}): Promise<Response> {
    // redirect:'manual' is essential — it turns the auth proxy's login redirect
    // into an opaqueredirect we can detect, instead of fetch auto-following it
    // cross-origin and throwing a CORS error we'd mistake for being offline.
    const res = await fetch(input, { ...init, redirect: "manual" });

    // An expired session looks different across auth proxies:
    //   • opaqueredirect — a 3xx login redirect, same- or cross-origin
    //     (tinyauth, Authelia, Authentik, Cloudflare Access…)
    //   • 401 / 403 — oauth2-proxy and friends reject XHR/JSON outright
    const isAuthChallenge =
        res.type === "opaqueredirect" ||
        res.status === 401 ||
        res.status === 403;

    if (isAuthChallenge) {
        // A top-level navigation lets the *browser* follow the redirect to the
        // login page (fetch can't, when it's cross-origin).
        if (!reloading) {
            reloading = true;
            window.location.assign(window.location.href);
        }
        throw new AuthRedirectError();
    }

    return res;
}
