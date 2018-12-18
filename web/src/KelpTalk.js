import axios from 'axios'
import env from './env.js'
import EventEmitter from 'events'

class KelpTalk extends EventEmitter {
  convertString(input) {
    let inString = input
    if (typeof input !== 'string') {
      inString = JSON.stringify(input, null, '  ')
    }

    if (inString && inString.length > 0) {
      return inString
      // return inString.replace(/\n/g, '<br>').replace(/\t/g, '  ')
    }

    return '<empty>'
  }

  get(path, params = {}) {
    axios.get(env.REST_Url + path, {params: params}).then(res => {
      let stringData = this.convertString(res.data)
      this.emit('console', stringData)

      switch (path) {
        case 'list':
          this.emit('updated', 'kelpTasks', res.data)
          break
        case 'version':
          // version is a string of lines
          // first line is version: xxxx
          const lines = res.data.split('\n')
          let vers = ''

          if (lines.length > 0) {
            // hack to shorten the string, maybe version will return json someday
            vers = lines[0].replace('version:', '').replace('master:', '').trim()
          }
          this.emit('updated', 'version', vers)
          break
        default:
          this.emit('updated', path, res.data)
          break
      }

    }).catch(err => {
      this.emit('console', JSON.stringify(err, null, '  '))

      console.error(err)
    });
  }

  put(path, params = {}) {
    axios.put(env.REST_Url + path, params).then(res => {
      let stringData = this.convertString(res.data)
      this.emit('console', stringData)
    }).catch(err => {
      this.emit('console', JSON.stringify(err, null, '  '))
      console.error(err)
    });
  }

  clearConsole() {
    this.emit('console', '')
  }
}

const instance = new KelpTalk()

export default instance
