/// <reference lib="webworker" />
import { clientsClaim } from 'workbox-core'
import { precacheAndRoute, getCacheKeyForURL } from 'workbox-precaching'
import { NavigationRoute, registerRoute } from 'workbox-routing'

declare const self: ServiceWorkerGlobalScope & { __WB_MANIFEST: any }

self.skipWaiting()
clientsClaim()

// Precache the built assets injected by vite-plugin-pwa.
//
// directoryIndex/cleanURLs are disabled so the precache route does NOT claim
// navigations to '/'. Otherwise it maps '/' -> '/index.html' and serves the
// cached shell straight from Cache Storage, bypassing the auth proxy entirely
// and shadowing the NavigationRoute below. With them off, '/' falls through to
// the NavigationRoute, which hits the network so the proxy's login redirect can
// reach the browser. The precached index.html stays available as the offline
// fallback via getCacheKeyForURL('/index.html').
precacheAndRoute(self.__WB_MANIFEST, {
  directoryIndex: '', // falsy → don't map '/' to '/index.html'
  cleanURLs: false,
})

// Navigations go to the network with redirect:'manual', so a cross-origin 302
// to the auth domain becomes an opaqueredirect the browser follows natively
// (no cached-shell short-circuit / reload loop). Fall back to the precached
// shell ONLY on a real network failure (genuinely offline).
registerRoute(
  new NavigationRoute(async ({ request }) => {
    try {
      return await fetch(request, { redirect: 'manual' })
    } catch {
      const cached = await caches.match(getCacheKeyForURL('/index.html') ?? '/index.html')
      return cached ?? Response.error()
    }
  }),
)
