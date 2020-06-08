import { createApp, callComponentsHookWith } from './app.js'

export default context => {
  return new Promise((resolve, reject) => {
    const { app, router,store } = createApp(context)

    router.push(context.url)
    router.onReady(() => {
      const matchedComponents = router.getMatchedComponents()
      if (!matchedComponents.length) {
        return reject({err:{code:404}, context})
      }

      callComponentsHookWith(matchedComponents, 'prepareVuex', { store })
      Promise.all(callComponentsHookWith(matchedComponents, 'asyncData',
          {store, route: router.currentRoute, context: context})
      ).then(() => {
        context.state = store.state
        context.meta = context.state.meta
        context.state.meta = {}

        resolve({app, context})
      }).catch((err) => {
        reject({err, context})
      });
    }, (err) => {
      reject({err, context})
    })
  })
}
