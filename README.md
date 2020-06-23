# gossr

gossr 是一个用于Web开发的服务器端渲染框架(SSR)，使用 golang + v8 实现，基于Vue搭建。类似于Nuxt，Next这类SSR框架，只是它们使用Nodejs实现。

## 优势
- SSR框架本身的优势。
  - 更好的SEO，搜索引擎爬虫可以直接抓取完全渲染的页面
  - 更快的内容到达时间 (time-to-content)，用户将会更快速地看到完整渲染的页面，从而有更好的用户体验。
- golang + v8 实现，相比基于Nodejs的方案，有更好的性能。
  - 使用golang实现服务器的框架，可以多线程调度多个V8 VM实例。
  - 并且在实际工程中，js往往会有内存泄漏，这个对于服务器是致命的，本框架通过设置V8 VM的生命期，生命期到后则删除该实例，从而解决了内存泄漏，保证服务器可以长时间稳定运行。
- 实现了SSR运行所需的js环境
  - 实现了CommonJS require加载规范
  - 实现了XMLHttpRequest，可以执行ajax请求
  - 实现了console.debug, console.log, console.info, console.warn, console.error
  - 没有实现SSR不推荐使用的setTimeout和setInterval方法，从而规避潜在的问题
