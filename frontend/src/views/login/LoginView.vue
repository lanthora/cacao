<template>
	<div class="login-container flex z-10">
		<div class="login-bg">
			<img src="@/assets/images/login_bg.jpg" />
		</div>
        <div class="login-con">
            <el-card class="login-card">
                <p class="login-title">欢迎登录Cacao平台</p>
                <div class="form-con">
                    <Account />
                </div>
            </el-card>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { defineAsyncComponent, onMounted, reactive, computed } from 'vue';

// 用户登录组件
const Account = defineAsyncComponent(() => import('@/views/login/component/account.vue'));

</script>

<style scoped>
.login-container {
	height: 100%;
	background: var(--el-color-white);

    .login-bg {
        display: flex;
        position: absolute;
        top: 0;
        left: 0;
        z-index: 1;
        animation: logoAnimation 0.3s ease;
        height: 100%;
        
        img {
            height: 100%;
            width: 100%;
        }
    }

    .login-con {
        position: absolute;
        top: 20%;
        right: 300px;
        width: 600px;
        height: 450px;
        z-index: 1;

        .login-card {
            width: 100%;
            height: 100%;

            .login-title {
                text-align: center;
                line-height: 56px;
                font-size: 30px;
                margin-top: 20px;
            }

            .form-con {
                margin-top: 30px;
                font-size: 26px;
            }
        }
    }
}
</style>