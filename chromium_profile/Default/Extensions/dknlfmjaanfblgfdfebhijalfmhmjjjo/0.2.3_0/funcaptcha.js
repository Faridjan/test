(async()=>{function u(){return null!==(document.querySelector('button[aria-describedby="descriptionVerify"]')||document.querySelector("#wrong_children_button")||document.querySelector("#wrongTimeout_children_button"))}function s(){try{var e=document.querySelector('button[aria-describedby="descriptionVerify"]'),t=(e&&(window.parent.postMessage({nopecha:!0,action:"clear"},"*"),e.click()),document.querySelector("#wrong_children_button")),a=(t&&(window.parent.postMessage({nopecha:!0,action:"clear"},"*"),t.click()),document.querySelector("#wrongTimeout_children_button"));a&&(window.parent.postMessage({nopecha:!0,action:"clear"},"*"),a.click())}catch(e){}}function d(){return document.querySelector("#game_children_text > h2")?.innerText?.trim()}function m(){return document.querySelector("img#game_challengeItem_image")?.src?.split(";base64,")[1]}let _=null;async function e(e){r=e,t=100;var r,t,{task:a,cells:n,image_data:c}=await new Promise(n=>{let c=!1;const i=setInterval(async()=>{if(!c){c=!0,r.funcaptcha_auto_open&&u()&&await s();var e=d();if(e){var t=document.querySelectorAll("#game_children_challenge ul > li > a");if(6===t.length){var a=m();if(a&&_!==a)return _=a,clearInterval(i),c=!1,n({task:e,cells:t,image_data:a})}}c=!1}},t)});if(null!==a&&null!==n&&null!==c){for(const l of[])if(a.startsWith(l))return;var i=Time.time(),o=(await NopeCHA.post({captcha_type:"funcaptcha",task:a,image_data:[c],key:e.key}))["data"];if(o){c=e.hcaptcha_solve_delay-(Time.time()-i);0<c&&await Time.sleep(c);for(let e=0;e<o.length;e++)!1!==o[e]&&n[e].click()}_=null}}if(window.location.pathname.startsWith("/fc/assets/tile-game-ui/"))for(;;){await Time.sleep(1e3);var t=await BG.exec("get_settings");t&&(t.funcaptcha_auto_open&&u()?await s():t.funcaptcha_auto_solve&&null!==d()&&null!==m()&&await e(t))}})();
