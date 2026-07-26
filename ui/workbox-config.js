module.exports = {
  globDirectory: "build/",
  globPatterns: [
    "**/*.{json,ico,html,png,txt,css,js}"
  ],
  swDest: "build/service-worker.js",
  skipWaiting: true,
  clientsClaim: true,
  runtimeCaching: [{
    // Logo 代理：在线命中后写入缓存，离线可继续显示图片
    // 必须放在通用 /api/ NetworkOnly 之前，优先匹配
    urlPattern: /\/api\/img(\?|$)/,
    handler: 'CacheFirst',
    options: {
      cacheName: 'van-nav-img',
      expiration: {
        maxEntries: 500,
        maxAgeSeconds: 30 * 24 * 60 * 60, // 30 days
      },
      cacheableResponse: {
        statuses: [0, 200],
      },
    },
  }, {
    urlPattern: /^https?.*\/api\//,
    handler: 'NetworkOnly',
  }, {
    urlPattern: /^https?.*/,
    handler: 'NetworkFirst',
    options: {
      cacheName: 'https-calls',
      networkTimeoutSeconds: 15,
      expiration: {
        maxEntries: 150,
        maxAgeSeconds: 30 * 24 * 60 * 60, // 30 days
      },
      cacheableResponse: {
        statuses: [0, 200],
      },
    },
  }]
};
