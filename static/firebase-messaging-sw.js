importScripts(
    'https://www.gstatic.com/firebasejs/8.6.2/firebase-app.js',
    'https://www.gstatic.com/firebasejs/8.6.2/firebase-analytics.js',
    'https://www.gstatic.com/firebasejs/8.6.2/firebase-messaging.js'
)
// Your web app's Firebase configuration
// For Firebase JS SDK v7.20.0 and later, measurementId is optional
const firebaseConfig = {
    apiKey: 'AIzaSyC0a8lgwVvetlH-o171X5taO7OzDtYL48M',
    authDomain: 'notification-dpl.firebaseapp.com',
    projectId: 'notification-dpl',
    storageBucket: 'notification-dpl.appspot.com',
    messagingSenderId: '9204240224',
    appId: '1:9204240224:web:f55113f81c66164ec91805',
    measurementId: 'G-DN4GG6GWS7',
}
firebase.initializeApp(firebaseConfig)

const messaging = firebase.messaging();

messaging.onBackgroundMessage((payload) => {
  console.log('[firebase-messaging-sw.js] Received background message ', payload);
  // Customize notification here
  // const notificationTitle = 'Background Message Title';
  // const notificationOptions = {
  //   body: 'Background Message body.',
  //   icon: '/firebase-logo.png'
  // };

  // self.registration.showNotification(notificationTitle,
  //   notificationOptions);
});
