
import os
import time
from datetime import datetime
from typing import Iterable, List
from zoneinfo import ZoneInfo

import boto3
import requests

_DISCORD_WEBHOOK_ID = os.environ['DISCORD_WEBHOOK_ID']
_DISCORD_WEBHOOK_TOKEN = os.environ['DISCORD_WEBHOOK_TOKEN']
_DISCORD_WEBHOOK_URL = 'https://discord.com/api/webhooks/{}/{}'.format(_DISCORD_WEBHOOK_ID, _DISCORD_WEBHOOK_TOKEN)

DYNAMODB_TABLE = os.environ.get('DYNAMODB_TABLE', 'showroom_timetable')
TIME_TABLE = 'https://www.showroom-live.com/api/time_table/time_tables'


class ShowroomProgramme:

    _url: str
    _member: str
    _group: str
    _room_id: int
    _start_at: int
    _created_at: int

    def __init__(self, record: dict) -> None:
        self._url = 'https://www.showroom-live.com/r/{}'.format(record['room_url_key'])
        self._member = record['member']
        self._group = record['group']
        self._room_id = record['room_id']
        self._start_at = record['started_at']
        self._created_at = record.get('created_at', int(time.time()))

    def to_dynamodb_format(self) -> dict:
        return {
            'url': {'S': self._url},
            'member': {'S': self._member},
            'group': {'S': self._group},
            'room_id': {'N': str(self._room_id)},
            'start_at': {'N': str(self._start_at)},
            'created_at': {'N': str(self._created_at)},
        }

    def to_json(self) -> dict:
        start = datetime.fromtimestamp(self._start_at, tz=ZoneInfo('Asia/Taipei'))
        return {
            'url': self._url,
            'member': self._member,
            'group': self._group,
            'room_id': self._room_id,
            'start_at': start.strftime('%Y-%m-%d %H:%M:%S'),
            'created_at': self._created_at,
        }


class DB:

    _db: boto3.client
    _table_name: str

    def __init__(self, table_name: str) -> None:
        self._table_name = table_name
        self._db = boto3.client('dynamodb')

    def set(self, showroom_program: ShowroomProgramme) -> bool:
        self._db.put_item(TableName=self._table_name, Item=showroom_program.to_dynamodb_format())

    def scan(self, start_from: int = 0) -> List[tuple]:
        resp = self._db.scan(
            TableName=self._table_name,
            FilterExpression='start_at > :start_from',
            ExpressionAttributeValues={':start_from': {'N': str(start_from)}})

        programmes = []
        for item in resp['Items']:
            programmes.append({
                'member': item['member']['S'],
                'group': item['group']
                ['S'], 'start_at': int(i['start_at']['N'])
            })
        return programmes


def get_latest_timetable() -> Iterable[dict]:
    resp = requests.get(TIME_TABLE, params={"order": 'asc', "started_at": int(time.time())}, timeout=3)
    print(resp.status_code)
    for slot in resp.json()['time_tables']:
        if '乃木坂46' in slot['main_name']:
            slot['member'], slot['group'] = slot['main_name'].replace(' ', '')[:-1].split('（')
            yield slot


def send_discord_message(message: ShowroomProgramme) -> None:
    _message = '''本日 Showroom 直播:
成員: {member}
時間: {start_at}
URL: {url}
'''
    resp = requests.post(
        url=_DISCORD_WEBHOOK_URL,
        json={'content': _message.format_map(message.to_json())},
        timeout=3,
    )
    print(resp.status_code)


if __name__ == '__main__':
    dynamodb = DB(table_name=DYNAMODB_TABLE)
    future_programmes = dynamodb.scan(int(time.time()))

    for i in get_latest_timetable():
        p = ShowroomProgramme(i)
        if not future_programmes:
            dynamodb.set(p)
            send_discord_message(p)
        else:
            for prog in future_programmes:
                if prog[0] == i['member'] and prog[1] == i['group'] and int(i['started_at']) > prog[2]:
                    dynamodb.set(p)
                    send_discord_message(p)
