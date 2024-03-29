
// cspell:words GOPACKAGE GOFILE

import { Parcel } from "@parcel/core"
import { mkdir, rm, writeFile, readFile, unlink, rmdir, } from "fs/promises"
import { join } from "path"
import { parse as parseHTML } from 'node-html-parser';

//
// PARAMETERS
//

const ENTRYPOINTS = process.argv.slice(2)
const ENTRY_DIR = join('.', '.entry-cache') // directory to place entries into
const DIST_DIR = join('.', 'dist')
const PUBLIC_DIR = '/assets/'

const DEST_PACKAGE = process.env.GOPACKAGE ?? 'static'
const DEST_FILE = (() => {
    const source = (process.env.GOFILE ?? 'assets.go')
    const base = source.substring(0, source.length - '.go'.length)
    return base + '_dist.go'
})()

//
// PREPARE DIRECTORIES
//

process.stdout.write('Preparing directories ...')
await Promise.all([
    mkdir(ENTRY_DIR, { recursive: true }),
    rm(DIST_DIR, { recursive: true, force: true })
])
console.log(' Done.')


//
// WRITE ENTRY POINTS
//

process.stdout.write('Collecting entry points ')
const entries = await Promise.all(ENTRYPOINTS.map(async (name) => {
    const entry = {
        'name': name,
        'bundleName': name + '.html',
        'src': join(ENTRY_DIR, name + '.html'),
    }

    const content = `
<script type='module' src='../src/base/index.ts'></script>
<script type='module' src='../src/entry/${name}/index.ts'></script>
<link rel='stylesheet' href='../src/entry/${name}/index.css'>
`;
    await writeFile(entry.src, content)

    process.stdout.write('.')
    return entry;
}))
console.log(' Done.')

//
// BUNDLEING
//

process.stdout.write('Bundleing assets ...')
const bundler = new Parcel({
    //env: process.env,
    entries: entries.map(e => e.src),
    defaultConfig: '@parcel/config-default',
    shouldDisableCache: true,
    shouldContentHash: true,
    defaultTargetOptions: {
        shouldOptimize: true,
        shouldScopeHoist: true,
        sourceMaps: false,
        distDir: DIST_DIR,
        publicUrl: PUBLIC_DIR,
        engines: {
            browsers: "defaults",
        }
    }
});
const { bundleGraph } = await bundler.run()
console.log(' Done.')


//
// FIND ASSETS IN OUTPUT
//

process.stdout.write('Find Assets in Output ')
const bundles = bundleGraph.getBundles()
const assets = await Promise.all(entries.map(async (entry) => {
    const mainBundle = bundles.find(b => b.name === entry.bundleName)
    if (mainBundle === undefined) throw new Error('Unable to find bundle for ' + entry.name)

    // read, then delete the generated output file
    const { filePath } = mainBundle
    const html = parseHTML(await readFile(filePath))
    await unlink(filePath)

    const scripts = html.querySelectorAll('script').map(script => script.outerHTML).join('')
    const links = html.querySelectorAll('link').map(link => link.outerHTML).join('')

    process.stdout.write('.')
    return { ...entry, scripts, links }
}))
console.log(' Done.')

//
// GENERATE GO
//

process.stdout.write(`Writing ${DEST_FILE} ...`)
const goAssets = assets.map(({ name, scripts, links }) => {
    return `
// Assets${name} contains assets for the '${name}' entrypoint.
var Assets${name} = Assets{
\tScripts: \`${scripts}\`,
\tStyles:  \`${links}\`,\t
}`.trim()
}).join('\n\n')
const goSource = `package ${DEST_PACKAGE}

// This file was automatically generated. Do not edit.
// cspell:disable

${goAssets}
`;

await writeFile(DEST_FILE, goSource)
console.log(' Done.')