import { inject as service } from '@ember/service';
import { alias } from '@ember/object/computed';
import { get, computed } from '@ember/object';
import Component from '@ember/component';
import decodeConfigFromJWT from 'replication/utils/decode-config-from-jwt';
import ReplicationActions from 'core/mixins/replication-actions';
import { task } from 'ember-concurrency';

const DEFAULTS = {
  token: null,
  id: null,
  loading: false,
  errors: [],
  primary_api_addr: null,
  primary_cluster_addr: null,
  ca_file: null,
  ca_path: null,
};

export default Component.extend(ReplicationActions, DEFAULTS, {
  replicationMode: 'dr',
  mode: 'primary',
  wizard: service(),
  version: service(),
  didReceiveAttrs() {
    this._super(...arguments);
    const initialReplicationMode = this.get('initialReplicationMode');
    if (initialReplicationMode) {
      this.set('replicationMode', initialReplicationMode);
    }
  },
  showModeSummary: false,
  initialReplicationMode: null,
  cluster: null,

  replicationAttrs: alias('cluster.replicationAttrs'),

  tokenIncludesAPIAddr: computed('token', function() {
    const config = decodeConfigFromJWT(get(this, 'token'));
    return config && config.addr ? true : false;
  }),

  disallowEnable: computed(
    'replicationMode',
    'version.hasPerfReplication',
    'mode',
    'tokenIncludesAPIAddr',
    'primary_api_addr',
    function() {
      const inculdesAPIAddr = this.get('tokenIncludesAPIAddr');
      if (this.get('replicationMode') === 'performance' && this.get('version.hasPerfReplication') === false) {
        return true;
      }
      if (
        this.get('mode') !== 'secondary' ||
        inculdesAPIAddr ||
        (!inculdesAPIAddr && this.get('primary_api_addr'))
      ) {
        return false;
      }
      return true;
    }
  ),

  reset() {
    this.setProperties(DEFAULTS);
  },

  submit: task(function*() {
    try {
      yield this.submitHandler.perform(...arguments);
    } catch (e) {
      // do not handle error
    }
  }),
  actions: {
    onSubmit(/*action, mode, data, event*/) {
      this.get('submit').perform(...arguments);
    },

    clear() {
      this.reset();
      this.setProperties({
        token: null,
        id: null,
      });
    },
  },
});
