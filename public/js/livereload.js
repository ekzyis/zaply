async function buildTime() {
  const r = await fetch("/__livereload", { cache: "no-cache"})
  return r.text()
}

async function liveReload() {
  console.log("running in development mode")

  let t_old = await buildTime()
  setInterval(async () => {
    let t_new = await buildTime()
    if (t_old !== t_new) {
      window.location.reload()
    }
  }, 1000)
}

liveReload().catch(console.error)