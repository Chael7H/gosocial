
// 好友界面Vue实例
new Vue({
    el: '#friends-app',
    data: {
        user: {
            id: '',
            nickname: '',
            avatar: ''
            // https://s3.bmp.ovh/imgs/2025/05/04/e272b0b155df44bd.png
        },
        friends: [],
        selectedFriend: null,
        newMessage: '',
        messages: [],
        searchQuery: '',
        searchResults: []
    },
    created() {
        this.loadUserInfo();
        this.loadFriendList();
        this.initWebSocket();
    },

    beforeDestroy() {
        if (this.socket) {
            this.socket.close();
        }
    },

    methods: {
        // 初始化WebSocket
        initWebSocket() {
            const token = localStorage.getItem('token');
            this.socket = new WebSocket(`ws://${window.location.host}/ws?token=${token}`);

            this.socket.onmessage = (event) => {
                const message = JSON.parse(event.data);
                switch(message.type) {
                    case 'new_message':
                        this.handleNewMessage(message.data);
                        break;
                    case 'unread_update':
                        this.updateUnreadCount(message.data.friend_id, message.data.count);
                        break;
                }
            };

            // 断线重连
            this.socket.onclose = () => {
                this.initWebSocket();
            };
        },

        // 处理新消息
        handleNewMessage(message) {
            // 只处理来自当前选中好友的消息
            if (this.selectedFriend && String(message.sender) === String(this.selectedFriend.id)) {
                this.processMessages([...this.messages, message]);
                this.scrollToBottom();

                // 只有接收到的消息才需要清空未读计数
                if (String(message.sender) !== String(this.user.id)) {
                    this.clearUnreadCount(message.sender);
                }
            }
            // 处理非当前会话的消息
            else if (String(message.sender) !== String(this.user.id)) {
                this.updateUnreadCount(message.sender, 1);
            }
        },

        // 更新未读计数
        updateUnreadCount(friendId, count) {
            const friend = this.friends.find(f => f.id === friendId);
            if (friend) {
                friend.unreadCount = (friend.unreadCount || 0) + count;
            }
        },

        // 清空未读计数
        clearUnreadCount(friendId) {
            const friend = this.friends.find(f => f.id === friendId);
            if (friend) {
                friend.unreadCount = 0;
            }
        },
        // 初始化WebSocket
        initWebSocket() {
            const token = localStorage.getItem('token');
            this.socket = new WebSocket(`ws://${window.location.host}/ws?token=${token}`);

            this.socket.onmessage = (event) => {
                const message = JSON.parse(event.data);
                if (message.type === 'new_message') {
                    this.handleNewMessage(message.data);
                }
            };
        },

        // 处理新消息
        handleNewMessage(message) {
            if (this.selectedFriend && message.sender === this.selectedFriend.id) {
                this.processMessages([...this.messages, message]);
                this.scrollToBottom();
                // 清空未读计数
                this.clearUnreadCount(message.sender);
            } else {
                this.updateUnreadCount(message.sender, 1);
            }
        },

        // 更新未读计数
        updateUnreadCount(friendId, count) {
            const friend = this.friends.find(f => f.id === friendId);
            if (friend) {
                friend.unreadCount = (friend.unreadCount || 0) + count;
            }
        },

        // 清空未读计数
        clearUnreadCount(friendId) {
            const friend = this.friends.find(f => f.id === friendId);
            if (friend) {
                friend.unreadCount = 0;
            }
        },

        // 加载用户信息
        loadUserInfo() {
            // 优先使用登录时存储的头像URL
            const loginAvatar = localStorage.getItem('avatar_url');

            // 设置默认头像路径
            const defaultAvatar = 'https://s3.bmp.ovh/imgs/2025/05/04/e272b0b155df44bd.png';

            axios.get('/api/v1/user/info', {
                headers: {
                    'Authorization': 'Bearer ' + localStorage.getItem('token')
                }
            })
                .then(response => {
                    const userId  = String(response.data.data.id);
                    const avatarUrl = loginAvatar || response.data.data.avatar || defaultAvatar;

                    // 预加载头像图片，确保可用
                    const img = new Image();
                    img.onload = () => {
                        this.user = {
                            id: userId,
                            nickname: response.data.data.nickname,
                            avatar: avatarUrl
                        };
                        localStorage.setItem('user_id', userId);
                        if (response.data.data.avatar) {
                            localStorage.setItem('avatar_url', response.data.data.avatar);
                        }
                    };
                    img.onerror = () => {
                        console.warn('头像加载失败，使用默认头像:', avatarUrl);
                        this.user = {
                            id: userId,
                            nickname: response.data.data.nickname,
                            avatar: defaultAvatar
                        };
                        localStorage.setItem('user_id', userId);
                    };
                    img.src = avatarUrl;
                })
                .catch(error => {
                    console.error('获取用户信息失败:', error);
                    // 使用本地存储的头像或默认头像
                    this.user = {
                        id: localStorage.getItem('user_id') || '',
                        nickname: '用户',
                        avatar: loginAvatar || defaultAvatar
                    };
                    alert('获取用户信息失败，正在使用缓存数据');
                });

            // 最终验证所有消息方向
            console.log('消息方向最终验证:');
            this.messages.forEach((msg, i) => {
                console.log(`消息${i}:`, {
                    content: msg.content.substring(0, 20),
                    from: msg.from,
                    to: msg.to,
                    direction: msg.isMe ? '右(我的消息)' : '左(好友消息)',
                    typeCheck: typeof msg.from === 'string' && typeof msg.to === 'string'
                });
            });
        },

        // 加载好友列表
        loadFriendList() {
            axios.get('/api/v1/friends', {
                headers: {
                    'Authorization': 'Bearer ' + localStorage.getItem('token')
                }
            })
                .then(response => {
                    if (response.data && response.data.data) {
                        this.friends = response.data.data.map(friend => {
                            return {
                                id: String(friend.friend_id || friend.id),
                                nickname: friend.display_name || friend.nickname,
                                remark: '',
                                avatar: friend.avatar_url || '/static/images/default-avatar.png',
                                lastMessage: '最近聊天',
                                lastTime: this.formatTime(friend.last_interact_at)
                            };
                        });
                    } else {
                        console.error('好友列表数据格式不正确:', response);
                    }
                })
                .catch(error => {
                    console.error('获取好友列表失败:', error);
                    alert('获取好友列表失败，请刷新重试');
                });
        },

        // 选择好友
        selectFriend(friend, event) {
            // 如果点击的是头像，则跳转到好友详情页
            if (event && event.target.classList.contains('friend-avatar')) {
                this.goToFriendDetail(friend.id);
                return;
            }

            this.selectedFriend = friend;
            this.searchQuery = '';
            this.searchResults = [];
            this.loadMessages(friend.id);
            // 清空未读计数
            this.clearUnreadCount(friend.id);
        },

        // 跳转到好友详情页
        goToFriendDetail(friendId) {
            window.open(`/friends/${friendId}`, '_blank');
        },

        // 加载消息记录
        loadMessages(friendId) {
            if (!friendId) return;

            axios.get('/api/v1/messages', {
                params: {
                    friend_id: friendId,
                    mark_read: true
                },
                headers: {
                    'Authorization': 'Bearer ' + localStorage.getItem('token')
                }
            })
                .then(response => {
                    console.log('完整响应:', response); // 调试日志
                    this.processMessages(response.data.data); // 正确访问嵌套数据
                    this.scrollToBottom();
                })
                .catch(error => {
                    console.error('获取消息失败:', error);
                });
        },

        // 加载历史消息
        loadHistoryMessages() {
            if (!this.selectedFriend) return;

            axios.get('/api/v1/messages', {
                params: {
                    friend_id: this.selectedFriend.id,
                    history: true,
                    mark_read: true
                },
                headers: {
                    'Authorization': 'Bearer ' + localStorage.getItem('token')
                }
            })
                .then(response => {
                    // 确保正确处理响应数据结构
                    const messages = response.data?.data?.messages || response.data?.messages || [];
                    this.processMessages(messages);
                    this.scrollToBottom();
                })
                .catch(error => {
                    console.error('获取历史消息失败:', error);
                    alert('获取历史消息失败，请重试');
                });
        },

        // 处理消息显示
        processMessages(response) {
            if (!this.user?.id) {
                console.error('用户ID未定义，无法处理消息');
                return;
            }
            // 确保ID始终为字符串类型
            const currentUserId = this.user.id;
            const messages = response?.messages || response?.data?.messages || [];

            console.log('用户ID验证:', {
                value: currentUserId,
                type: typeof currentUserId,
                expectedLength: currentUserId.length
            });

            this.messages = messages.map((msg, index) => {
                // 使用direct字段判断消息方向 (1=用户发给好友，2=好友发给用户)
                const isMe = msg.direct === 1;
                const hideTime = index > 0 &&
                    new Date(msg.created_at) - new Date(messages[index-1].created_at) < 5*60*1000;

                console.log('消息方向验证:', {
                    direct: msg.direct,
                    isMe: isMe,
                    expected: '1=用户发给好友, 2=好友发给用户'
                });

                // 确保当前用户消息使用自己的头像
                const avatarUrl = isMe
                    ? (this.user.avatar || 'https://s3.bmp.ovh/imgs/2025/05/04/e272b0b155df44bd.png')
                    : (msg.avatar_url || 'https://s3.bmp.ovh/imgs/2025/05/04/e272b0b155df44bd.png');

                // 统一处理消息内容
                const messageContent = msg.file_url || msg.content;
                return {
                    id: String(msg.id || index),
                    content: messageContent,
                    from: String(msg.from),
                    to: String(msg.to),
                    created_at: msg.created_at,
                    type: msg.type,
                    hide_time: hideTime,
                    isMe: isMe,
                    avatar_url: avatarUrl,
                    direct: msg.direct  // 保留direct字段用于调试
                };
            });
        },

        // 显示文件选择器
        showFilePicker(type) {
            const input = document.createElement('input');
            input.type = 'file';
            input.accept = type === 'image' ? 'image/*' : '*';
            input.onchange = (e) => {
                const file = e.target.files[0];
                if (!file) return;

                const formData = new FormData();
                formData.append('file', file);

                axios.post('/api/v1/upload', formData, {
                    headers: {
                        'Authorization': 'Bearer ' + localStorage.getItem('token'),
                        'Content-Type': 'multipart/form-data'
                    }
                })
                    .then(response => {
                        if (response.data && response.data.data && response.data.data.url) {
                            // 处理新响应格式
                            this.sendMessage(type === 'image' ? 2 : 3, response.data.data.url);
                        } else if (response.data && response.data.url) {
                            // 处理旧响应格式
                            this.sendMessage(type === 'image' ? 2 : 3, response.data.url);
                        } else {
                            throw new Error('上传成功但未返回文件URL');
                        }
                    })
                    .catch(error => {
                        console.error('上传失败:', error);
                        alert(`文件上传失败: ${error.response?.data?.message || error.message}`);
                    });
            };
            input.click();
        },

        // 发送消息
        sendMessage(type = 1, content = null) {  // 1:text, 2:image, 3:file
            if (!this.selectedFriend) return;

            const friendId = String(this.selectedFriend.id);
            const messageType = type ? parseInt(type) : 1;

            // 根据消息类型选择API路由
            const apiUrl = type === 2 ? '/api/v1/messages/image' :
                          type === 3 ? '/api/v1/messages/file' :
                          '/api/v1/messages';

            const message = {
                to: friendId,
                content: content || this.newMessage,
                ...(type === 2 ? {width: 0, height: 0} : {}),
                ...(type === 3 ? {name: '', size: 0, type: ''} : {})
            };

            axios.post(apiUrl, message, {
                headers: {
                    'Authorization': 'Bearer ' + localStorage.getItem('token'),
                    'Content-Type': 'application/json'
                }
            })
                .then(response => {
                    if (response.data.code === 1000) {
                        const newMsg = {
                            id: String(response.data.data.id),
                            from: String(this.user.id),
                            to: friendId,
                            content: message.content,
                            created_at: response.data.data.created_at || new Date().toISOString(),
                            type: messageType,
                            direct: 1,
                            isMe: true,
                            hide_time: false,
                            avatar_url: this.user.avatar
                        };

                        this.messages = [...this.messages, newMsg];
                        this.updateLastMessage(newMsg);
                        this.newMessage = ''; // 总是清空消息框
                        this.scrollToBottom();
                    } else {
                        alert(response.data.msg || '发送失败');
                    }
                })
                .catch(error => {
                    console.error('发送失败:', error);
                    alert(`消息发送失败: ${error.response?.data?.message || error.message}`);
                });
        },

        // 更新最后消息显示
        updateLastMessage(msg) {
            const friend = this.friends.find(f => f.id === this.selectedFriend.id);
            if (friend) {
                friend.lastMessage = msg.type === 'text' ? msg.content : `[${msg.type}]`;
                friend.lastTime = this.formatTime(msg.createdAt);
            }
        },

        // 跳转到用户信息页
        goToUserInfo() {
            const token = localStorage.getItem('token');
            const userInfoWindow = window.open('/user/info', '_blank');

            // 等待新窗口加载完成后传递token
            const checkLoaded = setInterval(() => {
                try {
                    userInfoWindow.postMessage({
                        type: 'SET_TOKEN',
                        token: token
                    }, window.location.origin);
                    clearInterval(checkLoaded);
                } catch (e) {
                    // 新窗口尚未准备好，继续等待
                }
            }, 100);
        },

        // 跳转到添加好友页
        goToAddFriend() {
            const friendId = prompt('请输入要添加的好友ID:');
            if (friendId) {
                axios.post(`/api/v1/friends/${friendId}`, {}, {
                    headers: {
                        'Authorization': 'Bearer ' + localStorage.getItem('token')
                    }
                })
                    .then(response => {
                        if (response.data.code === 1000) {
                            alert('好友添加成功!');
                            this.loadFriendList();
                        } else {
                            alert(response.data.msg || '添加失败');
                        }
                    })
                    .catch(error => {
                        if (error.response) {
                            const msg = error.response.data.msg ||
                                error.response.data.message ||
                                '添加失败';
                            alert(msg.includes('user not exist') ? '该用户不存在' : msg);
                        } else {
                            alert('网络错误，请重试');
                        }
                    });
            }
        },

        // 格式化简短时间
        formatTime(timestamp) {
            if (!timestamp) return '';
            const date = new Date(timestamp);
            return `${date.getHours()}:${date.getMinutes().toString().padStart(2, '0')}`;
        },

        // 格式化详细时间
        formatDetailedTime(timestamp) {
            if (!timestamp) return '';
            const date = new Date(timestamp);
            const today = new Date();
            const yesterday = new Date(today);
            yesterday.setDate(yesterday.getDate() - 1);

            if (date.toDateString() === today.toDateString()) {
                return `${date.getHours()}:${date.getMinutes().toString().padStart(2, '0')}`;
            } else if (date.toDateString() === yesterday.toDateString()) {
                return `昨天 ${date.getHours()}:${date.getMinutes().toString().padStart(2, '0')}`;
            } else {
                return `${date.getMonth()+1}月${date.getDate()}日 ${date.getHours()}:${date.getMinutes().toString().padStart(2, '0')}`;
            }
        },

        // 滚动到底部
        scrollToBottom() {
            this.$nextTick(() => {
                const container = document.querySelector('.messages-container');
                if (container) {
                    container.scrollTop = container.scrollHeight;
                }
            });
        },

        // 处理搜索输入
        handleSearchInput() {
            if (!this.searchQuery.trim()) {
                this.searchResults = [];
                return;
            }

            axios.get('/api/v1/friends/search', {
                params: {
                    keyword: this.searchQuery
                },
                headers: {
                    'Authorization': 'Bearer ' + localStorage.getItem('token')
                }
            })
                .then(response => {
                    this.searchResults = response.data.data.map(friend => ({
                        id: String(friend.friend_id),
                        nickname: friend.display_name,
                        avatar: friend.avatar_url || '/static/images/default-avatar.png'
                    }));
                })
                .catch(error => {
                    console.error('搜索好友失败:', error);
                    // 本地搜索作为fallback
                    const query = this.searchQuery.toLowerCase();
                    this.searchResults = this.friends.filter(friend =>
                        friend.nickname.toLowerCase().includes(query) ||
                        (friend.remark && friend.remark.toLowerCase().includes(query))
                    );
                });
        }
    }
});
