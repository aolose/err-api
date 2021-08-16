### Project struct
| database | ORM |  api | web |
|------------|--------|--------------|----|
|sqlit | gorm | iris | sveltejs kit|

### todo:
###### Functions:
- [x] Sign in and Sign out
- [ ] posts , tags , tag, page content
- [ ]  write a article   
    - [ ]  set public or draft  
    - [ ]  add res  
    - [ ]  pwd  
    - [ ]  seo field 
- [ ] comments 
    - [ ] create and list
    - [ ] firewall
- [ ] resources manage    
- [ ] front-end img compress
###### UI  
- [ ] front-end ui
###### Deploy
- [ ] move frontend to vercel 
- [ ] move backend to vultr 


#### What did I learn from this project:
The field 'Set-cookie' in response header  will affect your cookies via a fetch
function. I thought it's only work a doc request before. That means you can control 
your web's state via response. yes, you can't change status of your page via a fetch,
but cookie is working. Sometimes you may need render you content based on the 
cookie, just reload your page. If you change the session in sevlte kit,the location 
will reload, that's why the 'Set-cookie' can make you page change the content.
