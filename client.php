<?php



define('MAX_VALUE_LEN', 512);

define('STATUS_NOT_FOUND', 	4);
define('STATUS_NO', 		3);
define('STATUS_YES',		2);
define('STATUS_ERROR', 		1);

$data = packText("最近你还好吗？水电费水电费水电费水电费水电费水电费多少\sdfsdfsdf最近你还好吗？水电费水电费水电费水电费水电费水电费多少\sdfsdfsdf最近你还好吗？水电费水电费水电费水电费水电费水电费多少\sdfsdfsdf最近你还好吗？水电费水电费水电费水电费水电费水电费多少\sdfsdfsdf最近你还好吗？水电费水电费水电费水电费水电费水电费多少\sdfsdfsdf最近你还好吗？水电费水电费水电费水电费水电费水电费多少\sdfsdfsdf最近你还好吗？水电费水电费水电费水电费水电费水电费多少\sdfsdfsdf最近你还好吗？水电费水电费水电费水电费水电费水电费多少\sdfsdfsdf最近你还好吗？水电费水电费水电费水电费水电费水电费多少\sdfsdfsdf最近你还好吗？水电费水电费水电费水电费水电费水电费多少\sdfsdfsdf最近你还好吗？水电费水电费水电费水电费水电费水电费多少\sdfsdfsdf");
//$res = doRequest('/text/write', $data);

//var_dump($res);

$ret = doRequest('/text/get', array('textId' => '11571869116002780413'));
if ($ret['data']['flag'] == 1) {
	$ret['data']['text'] = gzuncompress($ret['data']['text']);
}
var_dump($ret);

function packText($text) {
	$len = strlen($text);
	$textId = getTextId($text);
	if ($textId < 0) {
		$textId = sprintf('%u', $textId);
	}
	$flag = 0;
	if ($len > 512) {
		$text = gzcompress($text, 7);
		$flag = 1;
	}
	return array(
		'textId' => $textId,
		'flag' => $flag,
		'text' => $text,
	);
}

function getTextId($text) {
	$hash = md5 ( $text, true );
	$high = substr ( $hash, 0, 8 );
	$low = substr ( $hash, 8, 8 );
	$sign = $high ^ $low;
	$sign1 = hexdec ( bin2hex ( substr ( $sign, 0, 4 ) ) );
	$sign2 = hexdec ( bin2hex ( substr ( $sign, 4, 4 ) ) );
	return ($sign1 << 32) | $sign2;
}


function doRequest($url, $data) {
	$fp = fsockopen('127.0.0.1', 8890, $errno, $errstr, 30);
	if (!$fp) {
		throw new Exception("doRequest.u_socketConnectFailed:$errstr {$errno}");
	} else {
		fwrite($fp, packRequest($url, $data));
		$response = false;
		if( !feof($fp) ) {
			$response = unpackResponse(fgets($fp, 1024));
		}	
		fclose($fp);
		return $response;
	}
}

function unpackResponse($str) {
	var_dump($str);
	$ret = array();
	$ret['status'] = array_shift(unpack('C', substr($str, 0, 1)));
	$ret['data'] = array();
	$str = substr($str, 1);
	while ($str) {
		if (substr($str, 0, 1) == "\0") {
			break;
		}
		$pos = strpos($str, "\0");
		if ($pos > -1) {
			$key = substr($str, 0, $pos);
			$str = substr($str, $pos + 1);

			$len = array_shift(unpack('n*', substr($str, 0, 2)));

			$value = substr($str, 2, $len);
			$ret['data'][$key] = $value;
			$str = substr($str, 2 + $len);
		}
	}
	return $ret;
}

function packRequest($url, array $data) {
	return $url."\0". packData($data);
}

function packData(array $data) {
	$ret = array();
	foreach($data as $key => $value) {
		$len = strlen($value);
		if ($len > MAX_VALUE_LEN) {
			throw new Exception('packData.u_outMaxRequestLen');
		}
		$ret[] = $key."\0".pack('n', $len).$value;
	}
	return implode('', $ret);
}