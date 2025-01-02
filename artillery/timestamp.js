let timestamp;

module.exports = {
  getTimestamp: (userContext, events, done) => {
    timestamp = new Date().getTime();
    userContext.vars.timestampValue = timestamp;
    return done();
  }
};
