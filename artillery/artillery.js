let timestamp;

module.exports = {
  getTimestamp: (userContext, events, done) => {
    timestamp = new Date().getTime();
    userContext.vars.timestampValue = timestamp;
    return done();
  },

  debugVariables: (userContext, events, done) => {
    console.log("==== DEBUG START ====");
    console.log(userContext.vars);
    console.log("Captured timestampValue:", userContext.vars.timestampValue); // 타임스탬프 값 출력
    console.log("Captured userEmail:", userContext.vars.userEmail);           // 이메일 값 출력
    console.log("Captured userId:", userContext.vars.userId);                 // 사용자 ID 값 출력
    console.log("Captured accessToken:", userContext.vars.accessToken);       // 토큰 값 출력
    console.log("==== DEBUG END ====");
    done();
  },

  extractAccessToken: (userContext, events, done) => {
    const headers = events.response.headers;
    const accessToken = headers['Authorization'];
    if (accessToken) {
      userContext.vars.accessToken = accessToken;
      console.log("Extracted accessToken:", userContext.vars.accessToken);
    } else {
      console.log("Authorization header not found in response");
    }
    return done();
  }
};
