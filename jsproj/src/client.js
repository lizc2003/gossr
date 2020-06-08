import { createApp, callComponentsHookWith } from './app.js'
import Router from 'vue-router'

const originRouterPush = Router.prototype.push
const originRouterReplace = Router.prototype.replace
Router.prototype.push = function push(location) {
  return originRouterPush.call(this, location).catch(err => err)
}
Router.prototype.replace = function replace(location) {
  return originRouterReplace.call(this, location).catch(err => err)
}

const { app, router,store } = createApp()

if (window.__INITIAL_STATE__) {
  store.replaceState(window.__INITIAL_STATE__)
}

router.onReady((initialRoute) => {
  const initialMatched = router.getMatchedComponents(initialRoute)
  callComponentsHookWith(initialMatched, 'prepareVuex', { store, isClientInitialRoute: true })

  router.beforeResolve((to, from, next) => {
    const matched = router.getMatchedComponents(to)
    const prevMatched = router.getMatchedComponents(from)

    let diffed = false
    const activated = matched.filter((c, i) => {
      return diffed || (diffed = (prevMatched[i] !== c))
    })
    if (!activated.length) {
      return next()
    }

    callComponentsHookWith(activated, 'prepareVuex', { store })
    Promise.all(callComponentsHookWith(activated, 'asyncData',
      { store, route: to, context:{}})
    ).then(() => {
      next()
    }).catch(next)
  })

  app.$mount('#app')
})
