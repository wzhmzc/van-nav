module.exports = {
  globDirectory: "build/",
  globPatterns: [
    "**/*.{json,ico,html,png,txt,css,js}"
  ],
  swDest: "build/service-worker.js",
  skipWaiting: true,
  clientsClaim: true,
  runtimeCaching: [{
    // Logo 代理：优先走网络，失败才回落缓存。
    // 原来是 CacheFirst + 30 天，一旦某次缓存到坏图（WAF 返回的 HTML 挑战页等），
    // 服务端修好了用户端仍会显示坏图长达一个月 —— 这类"修了看不见"的问题比多一次请求贵得多。
    urlPattern: /\/api\/img(\?|$)/,
    handler: 'NetworkFirst',
    options: {
      cacheName: 'van-nav-img',
      networkTimeoutSeconds: 3,
      expiration: {
        maxEntries: 500,
        maxAgeSeconds: 7 * 24 * 60 * 60, // 7 天（只作离线兜底）
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
