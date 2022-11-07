import{Cache,deep_copy,Time}from"./utils.mjs";import*as bapi from"./api.js";class Net{static async fetch({data:{url:t,options:e}}){try{return await(await fetch(t,e)).text()}catch(t){return null}}}class Tab{static reloads={};static _reload({tab_id:e}){return new Promise(t=>bapi.browser.tabs.reload(e,{bypassCache:!0},t))}static async reload({tab_id:t,data:{delay:e,overwrite:a}={delay:0,overwrite:!0}}){e=parseInt(e);var s=Tab.reloads[t]?.delay-(Date.now()-Tab.reloads[t]?.start),s=isNaN(s)||s<0?0:s;return!!(a||0==s||e<=s)&&(clearTimeout(Tab.reloads[t]?.timer),Tab.reloads[t]={delay:e,start:Date.now(),timer:setTimeout(()=>Tab._reload({tab_id:t}),e)},!0)}static close({tab_id:e}){return new Promise(t=>bapi.browser.tabs.remove(e,t))}static async open({data:{url:t}}){bapi.browser.tabs.create({url:t})}static info({tab_id:t}){return new Promise(e=>{try{bapi.browser.tabs.get(t,t=>e(t))}catch(t){e(!1)}})}}class Settings{static DEFAULT={version:4,key:"",hcaptcha_auto_solve:!0,hcaptcha_solve_delay:3e3,hcaptcha_auto_open:!0,recaptcha_auto_solve:!0,recaptcha_solve_delay:1e3,recaptcha_auto_open:!0,recaptcha_solve_method:"image",funcaptcha_auto_solve:!0,funcaptcha_solve_delay:1e3,funcaptcha_auto_open:!0,ocr_auto_solve:!1,ocr_image_selector:"",ocr_input_selector:"",debug:!1};static data={};static _save(){return new Promise(t=>bapi.browser.storage.sync.set({settings:Settings.data},t))}static load(){return new Promise(e=>{var t=bapi.browser.storage;t?t.sync.get(["settings"],async({settings:t})=>{t?(Settings.data=t,Settings.data.version!==Settings.DEFAULT.version&&(t=Settings.data.key,await Settings.reset(),Settings.data.key=t)):await Settings.reset(),Settings.data.key?.startsWith("MIIBI")&&(Settings.data.key=""),e()}):e()})}static async get(){return Settings.data}static async set({data:{id:t,value:e}}){Settings.data[t]=e,await Settings._save()}static async reset(){Settings.data=deep_copy(Settings.DEFAULT);var t=bapi.browser.runtime.getManifest();t.nopecha_key&&(Settings.data.key=t.nopecha_key),await Settings._save()}}class Injector{static inject({tab_id:t,data:{func:e,args:a}}){const s={target:{tabId:t,allFrames:!0},world:"MAIN",injectImmediately:!0,func:e,args:a};return new Promise(t=>bapi.browser.scripting.executeScript(s,t))}}class Recaptcha{static async reset({tab_id:t}){return await Injector.inject({tab_id:t,data:{func:function(){try{window.grecaptcha?.reset()}catch{}},args:[]}}),!0}}class Server{static ENDPOINT="https://api.nopecha.com/status?v="+bapi.browser.runtime.getManifest().version;static in_progress=!1;static async get_plan({data:{key:t}}){if(Server.in_progress)return!1;Server.in_progress=!0;let e={plan:"Unknown",credit:0};try{"undefined"===t&&(t="");var a=await fetch(Server.ENDPOINT+"&k="+t);e=JSON.parse(await a.text())}catch{}return Server.in_progress=!1,e}}const FN={set_cache:Cache.set,get_cache:Cache.get,remove_cache:Cache.remove,append_cache:Cache.append,empty_cache:Cache.empty,inc_cache:Cache.inc,dec_cache:Cache.dec,zero_cache:Cache.zero,fetch:Net.fetch,reload_tab:Tab.reload,close_tab:Tab.close,open_tab:Tab.open,info_tab:Tab.info,get_settings:Settings.get,set_settings:Settings.set,reset_settings:Settings.reset,reset_recaptcha:Recaptcha.reset,get_server_plan:Server.get_plan};(async()=>{bapi.register_language(),await Settings.load(),bapi.browser.runtime.onMessage.addListener((e,a,s)=>((async()=>{["get_settings","set_settings","set_cache"].includes(e.method);try{var t=await FN[e.method]({tab_id:a?.tab?.id,data:e.data});return t}catch(t){throw t}})().then(s).catch(t=>{s(t)}),!0))})();
