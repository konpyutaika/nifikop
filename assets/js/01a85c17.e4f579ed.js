"use strict";(self.webpackChunkreact_native_website=self.webpackChunkreact_native_website||[]).push([[8209],{86526:(e,t,a)=>{a.d(t,{in:()=>c,OU:()=>k,Ki:()=>A,kJ:()=>v,x:()=>l,e7:()=>u,J_:()=>f,Gx:()=>y});var n=a(63696),r=a(24613),i=a(96745),s=a(62540);function l(){var e,t=(0,i.A)(),a=null==t||null==(e=t.data)?void 0:e.blogMetadata;if(!a)throw new Error("useBlogMetadata() can't be called on the current route because the blog metadata could not be found in route context");return a}var o=n.createContext(null);function c(e){var t=e.children,a=e.content,r=e.isBlogPostPage,i=function(e){var t=e.content,a=e.isBlogPostPage;return(0,n.useMemo)((function(){return{metadata:t.metadata,frontMatter:t.frontMatter,assets:t.assets,toc:t.toc,isBlogPostPage:a}}),[t,a])}({content:a,isBlogPostPage:void 0!==r&&r});return(0,s.jsx)(o.Provider,{value:i,children:t})}function u(){var e=(0,n.useContext)(o);if(null===e)throw new r.dV("BlogPostProvider");return e}var d=a(10883),m=a(67032),g=function(e){return new Date(e).toISOString()};function h(e){var t=e.map(b);return{author:1===t.length?t[0]:t}}function p(e,t,a){return e?{image:(n={imageUrl:t(e,{absolute:!0}),caption:"title image for the blog post: "+a},r=n.imageUrl,i=n.caption,{"@type":"ImageObject","@id":r,url:r,contentUrl:r,caption:i})}:{};var n,r,i}function v(e){var t=(0,m.A)().siteConfig,a=(0,d.hH)().withBaseUrl,n=e.metadata,r=n.blogDescription,i=n.blogTitle,s=n.permalink,l=""+t.url+s;return{"@context":"https://schema.org","@type":"Blog","@id":l,mainEntityOfPage:l,headline:i,description:r,blogPost:e.items.map((function(e){return function(e,t,a){var n,r,i=e.assets,s=e.frontMatter,l=e.metadata,o=l.date,c=l.title,u=l.description,d=l.lastUpdatedAt,m=null!=(n=i.image)?n:s.image,v=null!=(r=s.keywords)?r:[],f=""+t.url+l.permalink,b=d?g(d):void 0;return Object.assign({"@type":"BlogPosting","@id":f,mainEntityOfPage:f,url:f,headline:c,name:c,description:u,datePublished:o},b?{dateModified:b}:{},h(l.authors),p(m,a,c),v?{keywords:v}:{})}(e.content,t,a)}))}}function f(){var e,t,a=l(),n=u(),r=n.assets,i=n.metadata,s=(0,m.A)().siteConfig,o=(0,d.hH)().withBaseUrl,c=i.date,v=i.title,f=i.description,b=i.frontMatter,j=i.lastUpdatedAt,x=null!=(e=r.image)?e:b.image,N=null!=(t=b.keywords)?t:[],C=j?g(j):void 0,y=""+s.url+i.permalink;return Object.assign({"@context":"https://schema.org","@type":"BlogPosting","@id":y,mainEntityOfPage:y,url:y,headline:v,name:v,description:f,datePublished:c},C?{dateModified:C}:{},h(i.authors),p(x,o,v),N?{keywords:N}:{},{isPartOf:{"@type":"Blog","@id":""+s.url+a.blogBasePath,name:a.blogTitle}})}function b(e){return Object.assign({"@type":"Person"},e.name?{name:e.name}:{},e.title?{description:e.title}:{},e.url?{url:e.url}:{},e.email?{email:e.email}:{},e.imageURL?{image:e.imageURL}:{})}var j=a(49519),x=a(45968),N=a(44544),C=a(34379);function y(e){var t=(0,j.zy)().pathname;return(0,n.useMemo)((function(){return e.filter((function(e){return function(e,t){return!(e.unlisted&&!(0,C.ys)(e.permalink,t))}(e,t)}))}),[e,t])}function A(e){var t=(0,N.$z)(e,(function(e){return""+new Date(e.date).getFullYear()})),a=Object.entries(t);return a.reverse(),a}function k(e){var t=e.items,a=e.ulClassName,n=e.liClassName,r=e.linkClassName,i=e.linkActiveClassName;return(0,s.jsx)("ul",{className:a,children:t.map((function(e){return(0,s.jsx)("li",{className:n,children:(0,s.jsx)(x.A,{isNavLink:!0,to:e.permalink,className:r,activeClassName:i,children:e.title})},e.permalink)}))})}},8109:(e,t,a)=>{a.d(t,{A:()=>H});var n=a(49257),r=a(63696),i=a(11750),s=a(6072),l=a(92575),o=a(6590),c=a(86526),u=a(27412),d=a(81381),m=a(62540);function g(e){var t=e.year,a=e.yearGroupHeadingClassName,n=e.children;return(0,m.jsxs)("div",{role:"group",children:[(0,m.jsx)(d.A,{as:"h3",className:a,children:t}),n]})}function h(e){var t=e.items,a=e.yearGroupHeadingClassName,n=e.ListComponent;if((0,u.p)().blog.sidebar.groupByYear){var r=(0,c.Ki)(t);return(0,m.jsx)(m.Fragment,{children:r.map((function(e){var t=e[0],r=e[1];return(0,m.jsx)(g,{year:t,yearGroupHeadingClassName:a,children:(0,m.jsx)(n,{items:r})},t)}))})}return(0,m.jsx)(n,{items:t})}const p=(0,r.memo)(h),v="sidebar_CLW8",f="sidebarItemTitle_DQfJ",b="sidebarItemList_rvuc",j="sidebarItem__RMN",x="sidebarItemLink_Ony9",N="sidebarItemLinkActive_MJ75",C="yearGroupHeading_jjLT";var y=function(e){var t=e.items;return(0,m.jsx)(c.OU,{items:t,ulClassName:(0,i.A)(b,"clean-list"),liClassName:j,linkClassName:x,linkActiveClassName:N})};function A(e){var t=e.sidebar,a=(0,c.Gx)(t.items);return(0,m.jsx)("aside",{className:"col col--3",children:(0,m.jsxs)("nav",{className:(0,i.A)(v,"thin-scrollbar"),"aria-label":(0,o.T)({id:"theme.blog.sidebar.navAriaLabel",message:"Blog recent posts navigation",description:"The ARIA label for recent posts in the blog sidebar"}),children:[(0,m.jsx)("div",{className:(0,i.A)(f,"margin-bottom--md"),children:t.title}),(0,m.jsx)(p,{items:a,ListComponent:y,yearGroupHeadingClassName:C})]})})}const k=(0,r.memo)(A);var _=a(65454);const P={yearGroupHeading:"yearGroupHeading_tAow"};var w=function(e){var t=e.items;return(0,m.jsx)(c.OU,{items:t,ulClassName:"menu__list",liClassName:"menu__list-item",linkClassName:"menu__link",linkActiveClassName:"menu__link--active"})};function O(e){var t=e.sidebar,a=(0,c.Gx)(t.items);return(0,m.jsx)(p,{items:a,ListComponent:w,yearGroupHeadingClassName:P.yearGroupHeading})}function G(e){return(0,m.jsx)(_.GX,{component:O,props:e})}const B=(0,r.memo)(G);function L(e){var t=e.sidebar,a=(0,l.l)();return null!=t&&t.items.length?"mobile"===a?(0,m.jsx)(B,{sidebar:t}):(0,m.jsx)(k,{sidebar:t}):null}var U=["sidebar","toc","children"];function H(e){var t=e.sidebar,a=e.toc,r=e.children,l=(0,n.A)(e,U),o=t&&t.items.length>0;return(0,m.jsx)(s.A,Object.assign({},l,{children:(0,m.jsx)("div",{className:"container margin-vert--lg",children:(0,m.jsxs)("div",{className:"row",children:[(0,m.jsx)(L,{sidebar:t}),(0,m.jsx)("main",{className:(0,i.A)("col",{"col--7":o,"col--9 col--offset-1":!o}),children:r}),a&&(0,m.jsx)("div",{className:"col col--2",children:a})]})})}))}},98968:(e,t,a)=>{a.r(t),a.d(t,{default:()=>v});a(63696);var n=a(11750),r=a(6590),i=function(){return(0,r.T)({id:"theme.tags.tagsPageTitle",message:"Tags",description:"The title of the tag list page"})};var s=a(18586),l=a(53237),o=a(8109),c=a(32515),u=a(81381);const d={tag:"tag_FHL6"};var m=a(62540);function g(e){var t=e.letterEntry;return(0,m.jsxs)("article",{children:[(0,m.jsx)(u.A,{as:"h2",id:t.letter,children:t.letter}),(0,m.jsx)("ul",{className:"padding--none",children:t.tags.map((function(e){return(0,m.jsx)("li",{className:d.tag,children:(0,m.jsx)(c.A,Object.assign({},e))},e.permalink)}))}),(0,m.jsx)("hr",{})]})}function h(e){var t=function(e){var t={};return Object.values(e).forEach((function(e){var a=function(e){return e[0].toUpperCase()}(e.label);null!=t[a]||(t[a]=[]),t[a].push(e)})),Object.entries(t).sort((function(e,t){var a=e[0],n=t[0];return a.localeCompare(n)})).map((function(e){return{letter:e[0],tags:e[1].sort((function(e,t){return e.label.localeCompare(t.label)}))}}))}(e.tags);return(0,m.jsx)("section",{className:"margin-vert--lg",children:t.map((function(e){return(0,m.jsx)(g,{letterEntry:e},e.letter)}))})}var p=a(689);function v(e){var t=e.tags,a=e.sidebar,r=i();return(0,m.jsxs)(s.e3,{className:(0,n.A)(l.G.wrapper.blogPages,l.G.page.blogTagsListPage),children:[(0,m.jsx)(s.be,{title:r}),(0,m.jsx)(p.A,{tag:"blog_tags_list"}),(0,m.jsxs)(o.A,{sidebar:a,children:[(0,m.jsx)(u.A,{as:"h1",children:r}),(0,m.jsx)(h,{tags:t})]})]})}},32515:(e,t,a)=>{a.d(t,{A:()=>l});a(63696);var n=a(11750),r=a(45968);const i={tag:"tag_otG2",tagRegular:"tagRegular_s0E1",tagWithCount:"tagWithCount_PGyn"};var s=a(62540);function l(e){var t=e.permalink,a=e.label,l=e.count,o=e.description;return(0,s.jsxs)(r.A,{href:t,title:o,className:(0,n.A)(i.tag,l?i.tagWithCount:i.tagRegular),children:[a,l&&(0,s.jsx)("span",{children:l})]})}}}]);