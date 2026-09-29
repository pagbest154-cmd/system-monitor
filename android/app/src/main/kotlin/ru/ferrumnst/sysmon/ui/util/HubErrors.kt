package ru.ferrumnst.sysmon.ui.util

import retrofit2.HttpException
import java.io.IOException
import java.net.ConnectException
import java.net.SocketTimeoutException
import java.net.UnknownHostException
import javax.net.ssl.SSLException

object HubErrors {
    fun isNetworkError(error: Throwable): Boolean {
        return when (error) {
            is IOException,
            is SocketTimeoutException,
            is ConnectException,
            is UnknownHostException,
            is SSLException -> true
            else -> error.cause?.let { isNetworkError(it) } == true
        }
    }

    fun userMessage(error: Throwable, fallback: String = "Ошибка"): String {
        if (error is HttpException && error.code() == 401) {
            return "Неверное имя хаба или ключ"
        }
        if (isNetworkError(error) || error is HttpException) {
            return "Хаб недоступен"
        }
        return error.message ?: fallback
    }
}
