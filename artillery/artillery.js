let timestamp;
let accessToken;

module.exports = {
  // 타임스탬프 생성
  getTimestamp: (userContext, events, done) => {
    timestamp = new Date().getTime();
    userContext.vars.timestampValue = timestamp;
    return done();
  },

  // 디버깅 변수 출력
  debugVariables: (userContext, events, done) => {
    console.log("==== DEBUG START ====");
    console.log("Captured variables:", userContext.vars); // 전체 변수 로그
    console.log("Captured timestampValue:", userContext.vars.timestampValue || "undefined");
    console.log("Captured userEmail:", userContext.vars.userEmail || "undefined");
    console.log("Captured userId:", userContext.vars.userId || "undefined");
    console.log("Captured accessToken:", userContext.vars.accessToken || "undefined");
    console.log("==== DEBUG END ====");
    done();
  },

  // Authorization 헤더에서 토큰 추출
  extractAuthToken: (userContext, events, done) => {
    const authHeader = userContext.vars.authHeader || "";
    if (authHeader.startsWith("Bearer ")) {
        userContext.vars.accessToken = authHeader.replace("Bearer ", ""); // 토큰 추출
        console.log("Extracted accessToken:", userContext.vars.accessToken);
    } else {
        console.log("Authorization header not found");
    }
    done();
  },
};
